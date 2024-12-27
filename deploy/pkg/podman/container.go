package podman

import (
	"fmt"
	"strings"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type Healthcheck struct {
	Test     []string
	Interval string
	Timeout  string
	Retries  int
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
	Healthcheck  *Healthcheck
	Networks     []string

	// Additional docker-compose style fields
	DNS         []string
	DNSSearch   []string
	Entrypoint  []string
	GroupAdd    []string
	Hostname    string
	Init        bool
	IpcMode     string
	Labels      map[string]string
	Links       []string
	LogConfig   LogConfig
	NetworkMode string
	PidMode     string
	Platform    string
	Privileged  bool
	ReadOnly    bool
	SecurityOpt []string
	ShmSize     string
	StopSignal  string
	StopTimeout *int
	Sysctls     map[string]string
	Tmpfs       []string
	Ulimits     map[string]UlimitConfig
	WorkingDir  string
	Pod         string
	DependsOn   map[string]DependsOnConfig `yaml:"depends_on,omitempty"`
}

type UlimitConfig struct {
	Soft int
	Hard int
}

type ContainerManager struct {
	ssh            *core.SSHConnection
	volumeManager  VolumeManagerInterface
	networkManager *NetworkManager
	maxRetries     int
	retryInterval  time.Duration
}

const (
	defaultMaxRetries    = 30
	defaultRetryInterval = 2 * time.Second
)

func NewContainerManager(ssh *core.SSHConnection) *ContainerManager {
	core.Logger.Debug("Creating new ContainerManager")
	return NewContainerManagerWithVolumes(ssh, NewVolumeManager(ssh))
}

func NewContainerManagerWithVolumes(ssh *core.SSHConnection, volumeManager VolumeManagerInterface) *ContainerManager {
	core.Logger.Debug("Creating new ContainerManager with custom VolumeManager")
	return &ContainerManager{
		ssh:            ssh,
		volumeManager:  volumeManager,
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

func (c *ContainerManager) waitForDependencies(container Container) error {
	core.Logger.Infof("Waiting for dependencies of container: %s", container.Name)
	for service, dep := range container.DependsOn {
		switch dep.Condition {
		case "service_healthy":
			if err := c.waitForHealthy(service); err != nil {
				core.Logger.Errorf("Dependency %s not healthy: %v", service, err)
				return err
			}
		case "service_started":
			if err := c.waitForRunning(service); err != nil {
				core.Logger.Errorf("Dependency %s not started: %v", service, err)
				return err
			}
		case "service_completed_successfully":
			if err := c.waitForSuccess(service); err != nil {
				core.Logger.Errorf("Dependency %s did not complete successfully: %v", service, err)
				return err
			}
		}
	}
	core.Logger.Infof("All dependencies for container %s are satisfied", container.Name)
	return nil
}

// Add new helper methods for dependency checking
func (c *ContainerManager) waitForRunning(name string) error {
	core.Logger.Infof("Waiting for container to reach running state: %s", name)
	for i := 0; i < c.maxRetries; i++ {
		output, err := c.ssh.Exec(fmt.Sprintf(
			"podman inspect --format '{{.State.Status}}' %s",
			name,
		))
		if err == nil {
			status := strings.TrimSpace(string(output))
			if status == "running" {
				core.Logger.Infof("Container is running: %s", name)
				return nil
			}
			// If container exited with error, no point in waiting
			if status == "exited" {
				core.Logger.Errorf("Container %s exited before reaching running state", name)
				return fmt.Errorf("container %s exited before reaching running state", name)
			}
		}
		time.Sleep(c.retryInterval)
	}
	core.Logger.Errorf("Container %s failed to reach running state within %v", name, time.Duration(c.maxRetries)*c.retryInterval)
	return fmt.Errorf("container %s failed to reach running state within %v", name, time.Duration(c.maxRetries)*c.retryInterval)
}

func (c *ContainerManager) waitForSuccess(name string) error {
	core.Logger.Infof("Waiting for container to complete successfully: %s", name)
	for i := 0; i < c.maxRetries; i++ {
		output, err := c.ssh.Exec(fmt.Sprintf(
			"podman inspect --format '{{.State.Status}}:{{.State.ExitCode}}' %s",
			name,
		))
		if err == nil {
			parts := strings.Split(strings.TrimSpace(string(output)), ":")
			if len(parts) == 2 {
				status := parts[0]
				exitCode := parts[1]

				// If container exited with 0, it completed successfully
				if status == "exited" && exitCode == "0" {
					core.Logger.Infof("Container completed successfully: %s", name)
					return nil
				}
				// If container exited with non-zero, it failed
				if status == "exited" && exitCode != "0" {
					core.Logger.Errorf("Container %s failed with exit code %s", name, exitCode)
					return fmt.Errorf("container %s failed with exit code %s", name, exitCode)
				}
			}
		}
		time.Sleep(c.retryInterval)
	}
	core.Logger.Errorf("Container %s did not complete within %v", name, time.Duration(c.maxRetries)*c.retryInterval)
	return fmt.Errorf("container %s did not complete within %v", name, time.Duration(c.maxRetries)*c.retryInterval)
}

func (c *ContainerManager) waitForHealthy(name string) error {
	core.Logger.Infof("Waiting for container to become healthy: %s", name)
	for i := 0; i < c.maxRetries; i++ {
		output, err := c.ssh.Exec(fmt.Sprintf(
			"podman inspect --format '{{.State.Health.Status}}' %s",
			name,
		))
		if err == nil && strings.TrimSpace(string(output)) == "healthy" {
			core.Logger.Infof("Container is healthy: %s", name)
			return nil
		}
		time.Sleep(c.retryInterval)
	}
	core.Logger.Errorf("Container failed to become healthy within %v", time.Duration(c.maxRetries)*c.retryInterval)
	return fmt.Errorf("container failed to become healthy within %v", time.Duration(c.maxRetries)*c.retryInterval)
}

func (c *ContainerManager) Remove(name string) error {
	core.Logger.Infof("Removing container: %s", name)
	_, err := c.ssh.Exec(fmt.Sprintf("podman rm -f %s", name))
	if err != nil {
		core.Logger.Errorf("Failed to remove container %s: %v", name, err)
		return err
	}
	return nil
}

func (c *ContainerManager) validateContainer(container Container) error {
	if container.Name == "" {
		return fmt.Errorf("container name is required")
	}
	if container.Image == "" {
		return fmt.Errorf("container image is required")
	}

	return nil
}

// Add Create method which is a simplified version of Run for initial container creation
func (c *ContainerManager) Create(container Container) error {
	core.Logger.Infof("Creating container: %s", container.Name)
	if err := c.validateContainer(container); err != nil {
		core.Logger.Errorf("Invalid container configuration: %v", err)
		return fmt.Errorf("invalid container configuration: %w", err)
	}

	// Handle dependencies first
	if err := c.waitForDependencies(container); err != nil {
		core.Logger.Errorf("Dependency check failed: %v", err)
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Use VolumeManager to handle volumes
	if err := c.volumeManager.EnsureVolumes(container.Volumes); err != nil {
		core.Logger.Errorf("Failed to ensure volumes: %v", err)
		return fmt.Errorf("failed to ensure volumes: %w", err)
	}

	cmd := strings.Builder{}
	cmd.WriteString(fmt.Sprintf("podman create --name %s", container.Name))

	// Add pod assignment if specified
	if container.Pod != "" {
		cmd.WriteString(fmt.Sprintf(" --pod %s", container.Pod))
	}

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
	_, err := c.ssh.Exec(cmd.String())
	if err != nil {
		core.Logger.Errorf("Failed to create container: %v", err)
		// Cleanup on failure
		if container.Name != "" {
			_, _ = c.ssh.Exec(fmt.Sprintf("podman rm -f %s", container.Name))
		}
		return fmt.Errorf("failed to create container: %w", err)
	}

	// Start the container
	_, err = c.ssh.Exec(fmt.Sprintf("podman start %s", container.Name))
	if err != nil {
		core.Logger.Errorf("Failed to start container: %v", err)
		// Cleanup on failure
		if container.Name != "" {
			_, _ = c.ssh.Exec(fmt.Sprintf("podman rm -f %s", container.Name))
		}
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Handle additional networks after container is created
	if len(container.Networks) > 1 {
		for _, network := range container.Networks[1:] {
			_, err := c.ssh.Exec(fmt.Sprintf(
				"podman network connect %s %s",
				network, container.Name,
			))
			if err != nil {
				core.Logger.Errorf("Failed to connect network %s: %v", network, err)
				return fmt.Errorf("failed to connect network %s: %w", network, err)
			}
		}
	}

	core.Logger.Infof("Container %s created and started successfully", container.Name)
	return nil
}

// MigrateContainer handles copying data between container versions
func (c *ContainerManager) MigrateContainer(serviceName, oldVersion, newVersion string, volumes []string) error {
	core.Logger.Info("Starting container data migration")

	for _, volume := range volumes {
		oldVolName := fmt.Sprintf("%s-%s-%s", serviceName, oldVersion, volume)
		newVolName := fmt.Sprintf("%s-%s-%s", serviceName, newVersion, volume)

		// Create new volume
		if err := c.volumeManager.EnsureVolume(newVolName); err != nil {
			return fmt.Errorf("failed to create new volume %s: %w", newVolName, err)
		}

		// Copy data from old to new volume using a temporary container
		if err := c.copyVolumeData(oldVolName, newVolName); err != nil {
			return fmt.Errorf("failed to copy volume data: %w", err)
		}

		core.Logger.Infof("Successfully migrated volume %s to %s", oldVolName, newVolName)
	}

	return nil
}

// copyVolumeData copies data between volumes using a temporary container
func (c *ContainerManager) copyVolumeData(sourceVol, destVol string) error {
	copyCmd := fmt.Sprintf(`podman run --rm \
		-v %s:/source:ro \
		-v %s:/destination \
		alpine sh -c 'cp -a /source/. /destination/'`,
		sourceVol, destVol)

	if _, err := c.ssh.Exec(copyCmd); err != nil {
		return fmt.Errorf("failed to copy data between volumes: %w", err)
	}

	return nil
}

type ContainerState struct {
	Health struct {
		Status string
	}
}

func (c *ContainerManager) Inspect(name string) (*ContainerState, error) {
	output, err := c.ssh.Exec(fmt.Sprintf("podman inspect --format '{{.State.Health.Status}}' %s", name))
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	return &ContainerState{
		Health: struct{ Status string }{
			Status: strings.TrimSpace(string(output)),
		},
	}, nil
}

type DependsOnConfig struct {
	Condition string `yaml:"condition,omitempty"` // service_started, service_healthy, service_completed_successfully
}
