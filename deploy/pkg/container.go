package pkg

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Healthcheck struct {
	Test     []string
	Interval string
	Timeout  string
	Retries  int
}

type DependencyConfig struct {
	Condition string // "service_started", "service_healthy", "service_completed_successfully"
	Service   string
}

type LogConfig struct {
	Driver  string
	Options map[string]string
}

type Container struct {
	Name         string
	Image        string
	Command      []string
	User         string
	Ports        []string
	Volumes      []string
	Environment  map[string]string
	EnvFile      []string
	Capabilities []string
	ExtraHosts   []string
	Restart      string
	Healthcheck  Healthcheck
	Networks     []string

	// Additional docker-compose style fields
	DependsOn    []DependencyConfig
	Deploy       DeployConfig
	DNS          []string
	DNSSearch    []string
	Entrypoint   []string
	GroupAdd     []string
	Hostname     string
	Init         bool
	IpcMode      string
	Labels       map[string]string
	Links        []string
	LogConfig    LogConfig
	NetworkMode  string
	PidMode      string
	Platform     string
	Privileged   bool
	ReadOnly     bool
	SecurityOpt  []string
	ShmSize      string
	StopSignal   string
	StopTimeout  *int
	Sysctls      map[string]string
	Tmpfs        []string
	Ulimits      map[string]UlimitConfig
	WorkingDir   string
}

type DeployConfig struct {
	Resources ResourceConfig
	Replicas  *int
	Labels    map[string]string
}

type ResourceConfig struct {
	Limits       Resources
	Reservations Resources
}

type Resources struct {
	CPUs    string
	Memory  string
	Devices []DeviceConfig
}

type DeviceConfig struct {
	Capabilities []string
	Count       *int
	Device      string
	Driver      string
	Options     map[string]string
}

type UlimitConfig struct {
	Soft int
	Hard int
}

type ContainerManager struct {
	ssh            SSHClientInterface
	volumeManager  VolumeManagerInterface
	networkManager *NetworkManager
	maxRetries     int
	retryInterval  time.Duration
}

const (
	defaultMaxRetries    = 30
	defaultRetryInterval = 2 * time.Second
)

func NewContainerManager(ssh SSHClientInterface) *ContainerManager {
	return &ContainerManager{
		ssh:            ssh,
		volumeManager:  NewVolumeManager(ssh),
		networkManager: NewNetworkManager(ssh),
		maxRetries:     defaultMaxRetries,
		retryInterval:  defaultRetryInterval,
	}
}

func (c *ContainerManager) WithTimeouts(maxRetries int, retryInterval time.Duration) *ContainerManager {
	c.maxRetries = maxRetries
	c.retryInterval = retryInterval
	return c
}

func (c *ContainerManager) Run(container Container) error {
	if err := c.validateContainer(container); err != nil {
		return fmt.Errorf("invalid container configuration: %w", err)
	}

	// Handle dependencies first
	if err := c.waitForDependencies(container); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Use VolumeManager to handle volumes
	if err := c.volumeManager.EnsureVolumes(container.Volumes); err != nil {
		return fmt.Errorf("failed to ensure volumes: %w", err)
	}

	cmd := strings.Builder{}
	cmd.WriteString(fmt.Sprintf("podman run -d --name %s", container.Name))

	// Basic configurations
	if container.Hostname != "" {
		cmd.WriteString(fmt.Sprintf(" --hostname %s", container.Hostname))
	}

	if container.User != "" {
		cmd.WriteString(fmt.Sprintf(" --user %s", container.User))
	}

	if container.WorkingDir != "" {
		cmd.WriteString(fmt.Sprintf(" --workdir %s", container.WorkingDir))
	}

	// Restart policy
	if container.Restart != "" {
		cmd.WriteString(fmt.Sprintf(" --restart %s", container.Restart))
	}

	// Capabilities
	for _, cap := range container.Capabilities {
		cmd.WriteString(fmt.Sprintf(" --cap-add %s", cap))
	}

	// Extra hosts
	for _, host := range container.ExtraHosts {
		cmd.WriteString(fmt.Sprintf(" --add-host %s", host))
	}

	// Ports
	for _, port := range container.Ports {
		cmd.WriteString(fmt.Sprintf(" -p %s", port))
	}

	// Volumes
	for _, volume := range container.Volumes {
		cmd.WriteString(fmt.Sprintf(" -v %s", volume))
	}

	// Environment files
	for _, envFile := range container.EnvFile {
		cmd.WriteString(fmt.Sprintf(" --env-file %s", envFile))
	}

	// Environment variables
	for key, value := range container.Environment {
		cmd.WriteString(fmt.Sprintf(" -e %s=%s", key, value))
	}

	// Healthcheck
	if len(container.Healthcheck.Test) > 0 {
		cmd.WriteString(fmt.Sprintf(" --health-cmd %q", strings.Join(container.Healthcheck.Test, " ")))
		if container.Healthcheck.Interval != "" {
			cmd.WriteString(fmt.Sprintf(" --health-interval %s", container.Healthcheck.Interval))
		}
		if container.Healthcheck.Timeout != "" {
			cmd.WriteString(fmt.Sprintf(" --health-timeout %s", container.Healthcheck.Timeout))
		}
		if container.Healthcheck.Retries > 0 {
			cmd.WriteString(fmt.Sprintf(" --health-retries %d", container.Healthcheck.Retries))
		}
	}

	// Network configurations
	if len(container.Networks) > 0 {
		cmd.WriteString(fmt.Sprintf(" --network %s", container.Networks[0]))
	}
	for _, dns := range container.DNS {
		cmd.WriteString(fmt.Sprintf(" --dns %s", dns))
	}
	for _, search := range container.DNSSearch {
		cmd.WriteString(fmt.Sprintf(" --dns-search %s", search))
	}

	// Resource limits
	if container.Deploy.Resources.Limits.CPUs != "" {
		cmd.WriteString(fmt.Sprintf(" --cpus %s", container.Deploy.Resources.Limits.CPUs))
	}
	if container.Deploy.Resources.Limits.Memory != "" {
		cmd.WriteString(fmt.Sprintf(" --memory %s", container.Deploy.Resources.Limits.Memory))
	}

	// Security configurations
	if container.Privileged {
		cmd.WriteString(" --privileged")
	}
	if container.ReadOnly {
		cmd.WriteString(" --read-only")
	}
	for _, secopt := range container.SecurityOpt {
		cmd.WriteString(fmt.Sprintf(" --security-opt %s", secopt))
	}

	// Labels
	for k, v := range container.Labels {
		cmd.WriteString(fmt.Sprintf(" --label %s=%s", k, v))
	}

	// Ulimits
	for name, ulimit := range container.Ulimits {
		cmd.WriteString(fmt.Sprintf(" --ulimit %s=%d:%d", name, ulimit.Soft, ulimit.Hard))
	}

	// Add the image
	cmd.WriteString(fmt.Sprintf(" %s", container.Image))

	// Add the command if specified
	if len(container.Command) > 0 {
		cmd.WriteString(fmt.Sprintf(" %s", strings.Join(container.Command, " ")))
	}

	// Execute the command
	execCmd := NewRemoteCommand(c.ssh, cmd.String())
	if err := execCmd.Run(); err != nil {
		// Cleanup on failure
		if container.Name != "" {
			cleanupCmd := NewRemoteCommand(c.ssh, fmt.Sprintf("podman rm -f %s", container.Name))
			_ = cleanupCmd.Run() // Ignore cleanup errors
		}
		return fmt.Errorf("failed to run container: %w", err)
	}

	// Handle additional networks after container is created
	if len(container.Networks) > 1 {
		for _, network := range container.Networks[1:] {
			connectCmd := NewRemoteCommand(c.ssh, fmt.Sprintf(
				"podman network connect %s %s",
				network, container.Name,
			))
			if err := connectCmd.Run(); err != nil {
				return fmt.Errorf("failed to connect network %s: %w", network, err)
			}
		}
	}

	return nil
}

func (c *ContainerManager) waitForDependencies(container Container) error {
	for _, dep := range container.DependsOn {
		switch dep.Condition {
		case "service_healthy":
			if err := c.waitForHealthy(dep.Service); err != nil {
				return err
			}
		case "service_started":
			if err := c.waitForRunning(dep.Service); err != nil {
				return err
			}
		case "service_completed_successfully":
			if err := c.waitForSuccess(dep.Service); err != nil {
				return err
			}
		}
	}
	return nil
}

// Add new helper methods for dependency checking
func (c *ContainerManager) waitForRunning(name string) error {
	for i := 0; i < c.maxRetries; i++ {
		cmd := NewRemoteCommand(c.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Status}}' %s",
			name,
		))
		output, err := cmd.Output()
		if err == nil {
			status := strings.TrimSpace(string(output))
			if status == "running" {
				return nil
			}
			// If container exited with error, no point in waiting
			if status == "exited" {
				return fmt.Errorf("container %s exited before reaching running state", name)
			}
		}
		time.Sleep(c.retryInterval)
	}
	return fmt.Errorf("container %s failed to reach running state within %v", name, time.Duration(c.maxRetries)*c.retryInterval)
}

func (c *ContainerManager) waitForSuccess(name string) error {
	for i := 0; i < c.maxRetries; i++ {
		cmd := NewRemoteCommand(c.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Status}}:{{.State.ExitCode}}' %s",
			name,
		))
		output, err := cmd.Output()
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(output)), ":")
			if len(parts) == 2 {
				status := parts[0]
				exitCode := parts[1]
				
				// If container exited with 0, it completed successfully
				if status == "exited" && exitCode == "0" {
					return nil
				}
				// If container exited with non-zero, it failed
				if status == "exited" && exitCode != "0" {
					return fmt.Errorf("container %s failed with exit code %s", name, exitCode)
				}
			}
		}
		time.Sleep(c.retryInterval)
	}
	return fmt.Errorf("container %s did not complete within %v", name, time.Duration(c.maxRetries)*c.retryInterval)
}

func (c *ContainerManager) waitForHealthy(name string) error {
	for i := 0; i < c.maxRetries; i++ {
		cmd := NewRemoteCommand(c.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Health.Status}}' %s",
			name,
		))
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "healthy" {
			return nil
		}
		time.Sleep(c.retryInterval)
	}
	return fmt.Errorf("container failed to become healthy within %v", time.Duration(c.maxRetries)*c.retryInterval)
}

func (c *ContainerManager) Remove(name string) error {
	cmd := NewRemoteCommand(c.ssh, fmt.Sprintf("podman rm -f %s", name))
	return cmd.Run()
}

func (c *ContainerManager) validateContainer(container Container) error {
	if container.Name == "" {
		return fmt.Errorf("container name is required")
	}
	if container.Image == "" {
		return fmt.Errorf("container image is required")
	}

	// Validate memory format
	if container.Deploy.Resources.Limits.Memory != "" {
		if !strings.HasSuffix(container.Deploy.Resources.Limits.Memory, "b") &&
		   !strings.HasSuffix(container.Deploy.Resources.Limits.Memory, "k") &&
		   !strings.HasSuffix(container.Deploy.Resources.Limits.Memory, "m") &&
		   !strings.HasSuffix(container.Deploy.Resources.Limits.Memory, "g") {
			return fmt.Errorf("invalid memory format: must end with b, k, m, or g")
		}
	}

	// Validate CPU format
	if container.Deploy.Resources.Limits.CPUs != "" {
		if _, err := strconv.ParseFloat(container.Deploy.Resources.Limits.CPUs, 64); err != nil {
			return fmt.Errorf("invalid CPU format: must be a number")
		}
	}

	return nil
}
