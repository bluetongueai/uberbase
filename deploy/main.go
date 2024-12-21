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
	Labels      map[string]string `yaml:"labels,omitempty"`
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
	config   DeployConfig
	ssh      pkg.SSHClientInterface
	workDir  string
	builder  *pkg.Builder
	registry *pkg.RegistryClient
	proxy    *pkg.ProxyManager
	state    *pkg.StateManager
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
	Service   string `yaml:"service"`
}

type StateManager interface {
	Load() (*State, error)
	Save(state *State) error
	GetLastDeployment() (*State, error)
	ClearState() error
}

type State struct {
	LastKnownGood map[string]Service
	Version       string
	Timestamp     time.Time
}

func NewDeployer(config DeployConfig) *Deployer {
	ssh := pkg.NewSSHClient(config.SSH.Host, config.SSH.User, config.SSH.Port, config.SSH.KeyFile, config.SSH.KeyData)
	workDir := "/tmp/deploy/" + filepath.Base(os.Getenv("PWD"))
	return &Deployer{
		config:   config,
		ssh:      ssh,
		workDir:  workDir,
		builder:  pkg.NewBuilder(ssh),
		registry: pkg.NewRegistryClient(ssh, config.Registry.Host, &pkg.RegistryAuth{
			Username: config.Registry.Username,
			Password: config.Registry.Password,
		}),
		proxy:    pkg.NewProxyManager(ssh, "./bin/kamal-proxy"),
		state:    pkg.NewStateManager(ssh, workDir),
	}
}

func (d *Deployer) Deploy() error {
	// 1. Connect via SSH
	if err := d.ssh.Connect(); err != nil {
		return fmt.Errorf("failed to connect SSH: %w", err)
	}
	defer d.ssh.Close()

	// 2. Git clone
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

	// 3. Build and push images first
	for name, service := range d.config.Services {
		if service.Build != nil {
			buildOpts := pkg.BuildOptions{
				File:         filepath.Join(d.workDir, service.Build.Dockerfile),
				ContextPath:  filepath.Join(d.workDir, service.Build.Context),
				Tag:          service.Image,
				BuildArgs:    service.Build.Args,
				Platform:     []string{"linux/amd64"},
				Pull:         "always",
				NetworkMode:  "host",
				Labels:       service.Labels,
				Memory:       service.Resources.Memory,
				CPUShares:    service.Resources.CPU,
				NoCache:      true,
				ForceRm:      true,
				SBOMOutput:   filepath.Join(d.workDir, "sbom.json"),
			}

			if err := d.builder.Build(buildOpts); err != nil {
				return fmt.Errorf("build failed for %s: %w", name, err)
			}

			if err := d.registry.PushImage(pkg.ImageRef{Name: service.Image}); err != nil {
				return fmt.Errorf("push failed for %s: %w", name, err)
			}
		}
	}

	// 4. Setup networks
	if err := d.setupNetworks(); err != nil {
		return fmt.Errorf("failed to setup networks: %w", err)
	}

	// 5. Setup volumes
	volManager := pkg.NewVolumeManager(d.ssh)
	for name := range d.config.Volumes {
		if err := volManager.EnsureVolume(name); err != nil {
			return fmt.Errorf("failed to setup volume %s: %w", name, err)
		}
	}

	// 6. Ensure proxy is running
	if err := d.ensureProxyRunning(); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	// 7. Load and save deployment state
	state, err := d.state.Load()
	if err != nil {
		return fmt.Errorf("failed to load deployment state: %w", err)
	}

	// Save current config as last known good
	state.LastKnownGood = d.config.Services
	if err := d.state.Save(state); err != nil {
		return fmt.Errorf("failed to save deployment state: %w", err)
	}

	// 8. Deploy services in order
	orderedServices := d.orderServices()
	for _, name := range orderedServices {
		service := d.config.Services[name]
		if err := d.deployService(name, service); err != nil {
			// 9. Rollback on failure
			log.Printf("Deployment failed for %s, rolling back to last known good configuration", name)
			for rollbackName, rollbackService := range state.LastKnownGood {
				if service, ok := rollbackService.(Service); ok {
					if err := d.deployService(rollbackName, service); err != nil {
						log.Printf("Rollback failed for %s: %v", rollbackName, err)
					}
				}
			}
			return fmt.Errorf("deployment failed for %s: %w", name, err)
		}
	}

	return nil
}

func (d *Deployer) deployService(name string, service Service) error {
	// 1. Check dependencies
	if service.DependsOn != nil {
		for dep, config := range service.DependsOn {
			if config.Condition == "service_healthy" {
				if err := d.waitForHealthy(dep); err != nil {
					return fmt.Errorf("dependency %s not healthy: %w", dep, err)
				}
			}
		}
	}

	// 2. Build and push if needed
	if service.Build != nil {
		buildOpts := pkg.BuildOptions{
			File:         filepath.Join(d.workDir, service.Build.Dockerfile),
			ContextPath:  filepath.Join(d.workDir, service.Build.Context),
			Tag:          service.Image,
			BuildArgs:    service.Build.Args,
			Platform:     []string{"linux/amd64"},
			Pull:         "always",
			NetworkMode:  "host",
			Labels:       service.Labels,
			Memory:       service.Resources.Memory,
			CPUShares:    service.Resources.CPU,
			NoCache:      true,
			ForceRm:      true,
			SBOMOutput:   filepath.Join(d.workDir, "sbom.json"),
		}

		if err := d.builder.Build(buildOpts); err != nil {
			return fmt.Errorf("build failed: %w", err)
		}

		if err := d.registry.PushImage(pkg.ImageRef{Name: service.Image}); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}
	}

	// 3. Start container first
	containerManager := pkg.NewContainerManager(d.ssh)
	container := pkg.Container{
		Name:         name,
		Image:        service.Image,
		Command:      d.parseCommand(service.Command),
		User:         service.User,
		Ports:        service.Ports,
		Volumes:      service.Volumes,
		Environment:  service.Environment,
		EnvFile:      service.EnvFile,
		Capabilities: service.Capabilities,
		ExtraHosts:   service.ExtraHosts,
		Restart:      service.Restart,
		Healthcheck: pkg.Healthcheck{
			Test:     service.Healthcheck.Test,
			Interval: service.Healthcheck.Interval,
			Timeout:  service.Healthcheck.Timeout,
			Retries:  service.Healthcheck.Retries,
		},
		Networks:    service.Networks,
		DependsOn:   make([]pkg.DependencyConfig, 0),
	}

	if service.DependsOn != nil {
		for svc, config := range service.DependsOn {
			container.DependsOn = append(container.DependsOn, pkg.DependencyConfig{
				Service:   svc,
				Condition: config.Condition,
			})
		}
	}

	// Connect container to networks before starting
	if len(service.Networks) > 0 {
		container.Networks = service.Networks
	}

	if err := containerManager.Run(container); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// 4. Configure proxy routing for each domain
	if len(service.Domains) > 0 {
		proxyService := pkg.ProxyService{
			Name:               name,
			Image:             service.Image,
			Domains:           service.Domains,
			SSL:               service.SSL,
			Networks:          service.Networks,
			Private:           service.Private,
			Port:              extractPort(service.Ports[0]),
			Environment:       service.Environment,
			Volumes:           service.Volumes,
			Command:           d.parseCommand(service.Command),
			Version:           "v1",
			Weight:            100,
			Labels:            service.Labels,
			HealthCheckTimeout: 30 * time.Second,
		}
		if err := validateProxyService(proxyService); err != nil {
			return fmt.Errorf("invalid proxy service configuration: %w", err)
		}
		if err := d.proxy.DeployService(proxyService); err != nil {
			return fmt.Errorf("proxy configuration failed: %w", err)
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

func (d *Deployer) resolveVolumes(volumes []string) ([]string, error) {
	volManager := pkg.NewVolumeManager(d.ssh)
	
	// Handle SELinux and volume options
	for i, vol := range volumes {
		options := strings.Split(vol, ":")
		if len(options) > 2 {
			if err := volManager.handleSELinux(options[0], options[2:]); err != nil {
				return nil, err
			}
			volumes[i] = strings.Join(volManager.handleVolumeOptions(options[0], options[2:]), ":")
		}
	}
	
	return volumes, nil
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

func extractPort(port string) string {
	parts := strings.Split(port, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func (d *Deployer) ensureProxyRunning() error {
	// Current implementation doesn't handle proxy installation
	// Should use the documented API:
	proxyInstaller := pkg.NewPodmanInstaller(d.ssh)
	if err := proxyInstaller.EnsureInstalled(); err != nil {
		return fmt.Errorf("failed to install proxy dependencies: %w", err)
	}
	
	// Check if proxy is running
	checkCmd := pkg.NewRemoteCommand(d.ssh, "pgrep kamal-proxy")
	if err := checkCmd.Run(); err != nil {
		// Start proxy if not running
		startCmd := pkg.NewRemoteCommand(d.ssh, "kamal-proxy run --http-port 80")
		if err := startCmd.Run(); err != nil {
			return fmt.Errorf("failed to start kamal-proxy: %w", err)
		}
	}
	return nil
}

func (d *Deployer) setupNetworks() error {
	netManager := pkg.NewNetworkManager(d.ssh)
	
	// List existing networks
	existing, err := netManager.ListNetworks()
	if err != nil {
		return err
	}
	
	// Create missing networks
	for name, net := range d.config.Networks {
		if !contains(existing, name) {
			if err := netManager.EnsureNetwork(name, net.Internal); err != nil {
				return fmt.Errorf("failed to create network %s: %w", name, err)
			}
		}
	}
	return nil
}

func validateProxyService(service pkg.ProxyService) error {
	if service.Name == "" {
		return fmt.Errorf("proxy service name is required")
	}
	if service.Image == "" {
		return fmt.Errorf("proxy service image is required")
	}
	if len(service.Domains) == 0 {
		return fmt.Errorf("proxy service domains are required")
	}
	if service.Port == "" {
		return fmt.Errorf("proxy service port is required")
	}
	return nil
}
