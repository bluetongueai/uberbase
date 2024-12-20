package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
	"github.com/bluetongueai/uberbase/deploy/pkg"
)

type ComposeConfig struct {
	Version  string              `yaml:"version,omitempty"`
	Services map[string]Service  `yaml:"services"`
	Volumes  map[string]Volume   `yaml:"volumes,omitempty"`
	Networks map[string]Network  `yaml:"networks,omitempty"`
}

type Service struct {
	Build       *BuildConfig         `yaml:"build,omitempty"`
	Image       string              `yaml:"image"`
	Command     interface{}         `yaml:"command,omitempty"` // string or []string
	User        string              `yaml:"user,omitempty"`
	Ports       []string            `yaml:"ports,omitempty"`
	Volumes     []string            `yaml:"volumes,omitempty"`
	Environment map[string]string   `yaml:"environment,omitempty"`
	EnvFile     []string           `yaml:"env_file,omitempty"`
	DependsOn   map[string]DependsOnConfig `yaml:"depends_on,omitempty"`
	CapAdd      []string            `yaml:"cap_add,omitempty"`
	ExtraHosts  []string            `yaml:"extra_hosts,omitempty"`
	Restart     string              `yaml:"restart,omitempty"`
	Healthcheck *Healthcheck        `yaml:"healthcheck,omitempty"`
	Networks    []string            `yaml:"networks,omitempty"`
	Capabilities []string         `yaml:"cap_add,omitempty"`
	Domains     []string         `yaml:"domains,omitempty"`
	SSL         bool             `yaml:"ssl,omitempty"`
	Private     bool             `yaml:"private,omitempty"`
}

type Volume struct {
	Driver string            `yaml:"driver,omitempty"`
	Name   string            `yaml:"name,omitempty"`
}

type Network struct {
	Driver string            `yaml:"driver,omitempty"`
	Internal bool           `yaml:"internal,omitempty"`
}

type BuildConfig struct {
	Context    string            `yaml:"context"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
}

type Deployer struct {
	config  DeployConfig
	ssh     *pkg.SSHClient
	workDir string
}

type DeployConfig struct {
	SSH      SSHConfig
	Registry RegistryConfig
	Services map[string]Service  // Direct access to services
	Networks map[string]Network
	Volumes  map[string]Volume
}

type SSHConfig struct {
	Host     string
	User     string
	Port     int
	KeyFile  string
	KeyData  string
}

type RegistryConfig struct {
	Host     string
	Username string
	Password string
}

type Healthcheck struct {
	Test     []string `yaml:"test"`
	Interval string   `yaml:"interval"`
	Timeout  string   `yaml:"timeout"`
	Retries  int      `yaml:"retries"`
}

type DependsOnConfig struct {
	Condition string `yaml:"condition"`
}

func NewDeployer(config DeployConfig) *Deployer {
	ssh := pkg.NewSSHClient(config.SSH.Host, config.SSH.User, config.SSH.Port, config.SSH.KeyFile, config.SSH.KeyData)
	return &Deployer{
		config:  config,
		ssh:     ssh,
		workDir: "/tmp/deploy/" + filepath.Base(os.Getenv("PWD")),
	}
}

func (d *Deployer) Deploy() error {
	if err := d.ssh.Connect(); err != nil {
		return fmt.Errorf("failed to connect SSH: %w", err)
	}
	defer d.ssh.Close()

	// Setup networks
	netManager := pkg.NewNetworkManager(d.ssh)
	for name, net := range d.config.Networks {
		if err := netManager.EnsureNetwork(name, net.Internal); err != nil {
			return fmt.Errorf("failed to setup network %s: %w", name, err)
		}
	}

	// Setup volumes
	volManager := pkg.NewVolumeManager(d.ssh)
	for name := range d.config.Volumes {
		if err := volManager.EnsureVolume(name); err != nil {
			return fmt.Errorf("failed to setup volume %s: %w", name, err)
		}
	}

	// Setup Git repository
	gitURL, err := d.getCurrentRepoURL()
	if err != nil {
		return fmt.Errorf("failed to get git URL: %w", err)
	}

	cloner := pkg.NewGitCloner(d.ssh)
	if err := cloner.Clone(pkg.CloneOptions{
		URL:         gitURL,
		Destination: d.workDir,
		Depth:       1,
	}); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Setup registry client
	registry := pkg.NewRegistryClient(d.ssh, d.config.Registry.Host, &pkg.RegistryAuth{
		Username: d.config.Registry.Username,
		Password: d.config.Registry.Password,
	})

	// Setup builder
	builder := pkg.NewBuilder(d.ssh)

	// Setup proxy manager
	proxy := pkg.NewProxyManager(d.ssh, "/usr/local/bin/kamal-proxy")

	// Order services by dependencies
	orderedServices := d.orderServices()

	// Process each service
	for _, name := range orderedServices {
		service := d.config.Services[name]
		if err := d.deployService(name, service, builder, registry, proxy); err != nil {
			return fmt.Errorf("failed to deploy service %s: %w", name, err)
		}
	}

	return nil
}

func (d *Deployer) deployService(name string, service Service, builder *pkg.Builder, registry *pkg.RegistryClient, proxy *pkg.ProxyManager) error {
	// Check dependencies
	for dep, config := range service.DependsOn {
		// Verify dependency exists
		_, exists := d.config.Services[dep]
		if !exists {
			return fmt.Errorf("dependency %s not found", dep)
		}

		// Handle service_healthy condition
		if config.Condition == "service_healthy" {
			if err := d.waitForHealthy(dep); err != nil {
				return fmt.Errorf("dependency %s failed health check: %w", dep, err)
			}
		}
	}

	// Ensure volumes exist
	if len(service.Volumes) > 0 {
		volManager := pkg.NewVolumeManager(d.ssh)
		if err := volManager.EnsureVolumes(service.Volumes); err != nil {
			return fmt.Errorf("failed to setup volumes: %w", err)
		}
	}

	// Build the image
	buildOpts := pkg.BuildOptions{
		Dockerfile:  filepath.Join(d.workDir, service.Build.Dockerfile),
		ContextPath: filepath.Join(d.workDir, service.Build.Context),
		Tag:        service.Image,
		BuildArgs:  service.Build.Args,
		Environment: service.Environment,
		Volumes:    service.Volumes,
	}

	if err := builder.Build(buildOpts); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Handle healthcheck if defined
	if len(service.Healthcheck.Test) > 0 {
		cmd := strings.Join(service.Healthcheck.Test, " ")
		healthCmd := pkg.NewRemoteCommand(d.ssh, cmd)
		if err := healthCmd.Run(); err != nil {
			return fmt.Errorf("healthcheck failed: %w", err)
		}
	}

	// Push to registry
	imageRef := pkg.ImageRef{
		Name: service.Image,
	}
	if err := registry.PushImage(imageRef); err != nil {
		return fmt.Errorf("push failed: %w", err)
	}

	// Create container
	container := pkg.Container{
		Name:        name,
		Image:       service.Image,
		Command:     d.parseCommand(service.Command),
		User:        service.User,
		Ports:       service.Ports,
		Volumes:     service.Volumes,
		Environment: service.Environment,
		EnvFile:     service.EnvFile,
		Capabilities: service.Capabilities,
		ExtraHosts:  service.ExtraHosts,
		Restart:     service.Restart,
		Healthcheck: pkg.Healthcheck{
			Test:     service.Healthcheck.Test,
			Interval: service.Healthcheck.Interval,
			Timeout:  service.Healthcheck.Timeout,
			Retries:  service.Healthcheck.Retries,
		},
	}

	// Create and start container
	containerManager := pkg.NewContainerManager(d.ssh)
	if err := containerManager.Run(container); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Connect container to networks
	netManager := pkg.NewNetworkManager(d.ssh)
	if err := netManager.ConnectContainer(name, service.Networks); err != nil {
		return fmt.Errorf("failed to connect networks: %w", err)
	}

	// Configure proxy
	proxyService := pkg.ProxyService{
		Name:    name,
		Domains: service.Domains,
		Port:    service.Ports[0],
		SSL:     service.SSL,
		Networks: service.Networks,
	}

	if err := proxy.DeployService(proxyService); err != nil {
		return fmt.Errorf("proxy configuration failed: %w", err)
	}

	// Connect to appropriate network
	if service.Private {
		if err := netManager.ConnectContainer(name, []string{"private"}); err != nil {
			return fmt.Errorf("failed to connect to private network: %w", err)
		}
	} else {
		if err := netManager.ConnectContainer(name, []string{"public"}); err != nil {
			return fmt.Errorf("failed to connect to public network: %w", err)
		}
	}

	return nil
}

func (d *Deployer) getCurrentRepoURL() (string, error) {
	output, err := runLocalCommand("git config --get remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func runLocalCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

const usage = `Usage: deploy [options]

A tool to deploy services from docker-compose.yml using kamal-proxy.

Required:
  --ssh-host <host>      SSH host to deploy to

Optional:
  -h, --help              Show this help message
  -f <file>              Path to docker-compose.yml (default: docker-compose.yml)
  --ssh-user <user>      SSH user (default: root)
  --ssh-port <port>      SSH port (default: 22)
  -i <keyfile>           SSH private key file
  --ssh-key-env <name>   Environment variable containing SSH key (default: SSH_PRIVATE_KEY)
  --registry <url>       Registry URL (default: docker.io)
  --registry-user <user> Registry username
  --registry-pass <pass> Registry password

Examples:
  # Using SSH key file
  deploy --ssh-host prod.example.com --ssh-user deploy -i ~/.ssh/prod_key

  # Using SSH key from environment
  SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" deploy --ssh-host prod.example.com --ssh-user deploy

  # Using custom compose file
  deploy --ssh-host prod.example.com --ssh-user deploy -f docker-compose.prod.yml

  # Minimal usage
  deploy --ssh-host prod.example.com
`

func main() {
	// Set custom usage before defining flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	var (
		sshHost     = flag.String("ssh-host", "", "SSH host")
		sshUser     = flag.String("ssh-user", "root", "SSH user")
		sshPort     = flag.Int("ssh-port", 22, "SSH port")
		sshKeyFile  = flag.String("i", "", "SSH private key file")
		sshKeyEnv   = flag.String("ssh-key-env", "SSH_PRIVATE_KEY", "Environment variable containing SSH key")
		registryURL = flag.String("registry", "docker.io", "Registry URL")
		regUser     = flag.String("registry-user", "", "Registry username")
		regPass     = flag.String("registry-pass", "", "Registry password")
		composePath = flag.String("f", "docker-compose.yml", "Path to docker-compose.yml")
	)

	flag.Parse()

	// Validate required flags
	if *sshHost == "" {
		log.Fatal("SSH host and user are required")
	}

	// Read docker-compose.yml
	data, err := os.ReadFile(*composePath)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", *composePath, err)
	}

	var config ComposeConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Fatalf("Failed to parse %s: %v", *composePath, err)
	}

	// Get SSH key from environment if specified
	var sshKeyData string
	if *sshKeyFile == "" {
		sshKeyData = os.Getenv(*sshKeyEnv)
		if sshKeyData == "" {
			sshKeyData = os.Getenv("SSH_PRIVATE_KEY")
		}
	}

	deployer := NewDeployer(DeployConfig{
		SSH: SSHConfig{
			Host: *sshHost,
			User: *sshUser,
			Port: *sshPort,
			KeyFile: *sshKeyFile,
			KeyData: sshKeyData,
		},
		Registry: RegistryConfig{
			Host:     *registryURL,
			Username: *regUser,
			Password: *regPass,
		},
		Services: config.Services,
		Networks: config.Networks,
		Volumes:  config.Volumes,
	})

	if err := deployer.Deploy(); err != nil {
		log.Fatalf("Deployment failed: %v", err)
	}
}

func (d *Deployer) orderServices() []string {
	visited := make(map[string]bool)
	var order []string

	var visit func(name string)
	visit = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		
		service := d.config.Services[name]
		for dep := range service.DependsOn {
			visit(dep)
		}
		order = append(order, name)
	}

	for name := range d.config.Services {
		visit(name)
	}

	return order
}

func (d *Deployer) resolveVolumes(volumes []string) []string {
	resolved := make([]string, len(volumes))
	for i, vol := range volumes {
		if strings.Contains(vol, ":") {
			resolved[i] = os.ExpandEnv(vol)
			continue
		}
		// Handle named volumes
		if _, exists := d.config.Volumes[vol]; exists {
			resolved[i] = vol + ":/" + vol
		}
	}
	return resolved
}

func (d *Deployer) waitForHealthy(serviceName string) error {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		cmd := pkg.NewRemoteCommand(d.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Health.Status}}' %s",
			serviceName,
		))
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "healthy" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("service %s failed to become healthy", serviceName)
}

func (d *Deployer) parseCommand(cmd interface{}) []string {
	switch v := cmd.(type) {
	case string:
		return strings.Fields(v)
	case []string:
		return v
	default:
		return nil
	}
}
