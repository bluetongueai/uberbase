package pkg

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type VolumeManager struct {
	ssh *SSHClient
}

func NewVolumeManager(ssh *SSHClient) *VolumeManager {
	return &VolumeManager{
		ssh: ssh,
	}
}

func (v *VolumeManager) EnsureVolume(name string) error {
	// Check if volume exists
	cmd := NewRemoteCommand(v.ssh, fmt.Sprintf("podman volume ls | grep -q %s", name))
	if err := cmd.Run(); err == nil {
		return nil // Volume exists
	}

	// Create volume
	cmd = NewRemoteCommand(v.ssh, fmt.Sprintf("podman volume create %s", name))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create volume %s: %w", name, err)
	}

	return nil
}

func (v *VolumeManager) handleSELinux(hostPath string, options []string) error {
	// Default to private label if Z is specified
	private := false
	shared := false

	for _, opt := range options {
		switch opt {
		case "Z":
			private = true
		case "z":
			shared = true
		}
	}

	if !private && !shared {
		return nil
	}

	// Build semanage command
	var cmd *RemoteCommand
	if private {
		// -t container_file_t for private container content
		cmd = NewRemoteCommand(v.ssh, fmt.Sprintf(
			"chcon -Rt container_file_t %s",
			hostPath,
		))
	} else {
		// -t container_share_t for shared container content
		cmd = NewRemoteCommand(v.ssh, fmt.Sprintf(
			"chcon -Rt container_share_t %s",
			hostPath,
		))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set SELinux context on %s: %w", hostPath, err)
	}

	return nil
}

func (v *VolumeManager) handleVolumeOptions(hostPath string, options []string) []string {
	var mountOpts []string
	var selinuxOpts []string

	for _, opt := range options {
		switch opt {
		// SELinux
		case "z", "Z":
			selinuxOpts = append(selinuxOpts, opt)
		
		// Access modes
		case "ro", "rw":
			mountOpts = append(mountOpts, opt)
			
		// Bind propagation
		case "shared", "slave", "private",
			 "rshared", "rslave", "rprivate":
			mountOpts = append(mountOpts, opt)
			
		// Other options
		case "nocopy":
			mountOpts = append(mountOpts, opt)
		case "consistent", "cached", "delegated":
			mountOpts = append(mountOpts, opt)
		}
	}

	// Handle SELinux separately
	if len(selinuxOpts) > 0 {
		if err := v.handleSELinux(hostPath, selinuxOpts); err != nil {
			// Log error but continue - SELinux might not be enabled
			fmt.Printf("Warning: SELinux labeling failed: %v\n", err)
		}
		// Add SELinux options to mount options for podman
		mountOpts = append(mountOpts, selinuxOpts...)
	}

	return mountOpts
}

func (v *VolumeManager) EnsureVolumes(volumes []string) error {
	for _, volume := range volumes {
		if strings.Contains(volume, ":") {
			parts := strings.Split(volume, ":")
			hostPath := os.ExpandEnv(parts[0])

			// Create host directory
			cmd := NewRemoteCommand(v.ssh, fmt.Sprintf("mkdir -p %s", hostPath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to create bind mount directory %s: %w", hostPath, err)
			}

			// Handle mount options
			if len(parts) > 2 {
				options := strings.Split(parts[2], ",")
				_ = v.handleVolumeOptions(hostPath, options)
			}
		} else {
			// Named volume
			if err := v.EnsureVolume(volume); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VolumeManager) RemoveVolume(name string) error {
	cmd := NewRemoteCommand(v.ssh, fmt.Sprintf("podman volume rm %s", name))
	return cmd.Run()
}

func (v *VolumeManager) ListVolumes() ([]string, error) {
	cmd := NewRemoteCommand(v.ssh, "podman volume ls --format '{{.Name}}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	volumes := strings.Split(strings.TrimSpace(string(output)), "\n")
	return volumes, nil
}

type NetworkManager struct {
	ssh *SSHClient
}

func NewNetworkManager(ssh *SSHClient) *NetworkManager {
	return &NetworkManager{
		ssh: ssh,
	}
}

func (n *NetworkManager) EnsureNetwork(name string, internal bool) error {
	// Check if network exists
	cmd := NewRemoteCommand(n.ssh, fmt.Sprintf("podman network ls | grep -q %s", name))
	if err := cmd.Run(); err == nil {
		return nil // Network already exists
	}

	// Build network create command
	createCmd := strings.Builder{}
	createCmd.WriteString(fmt.Sprintf("podman network create %s", name))

	if internal {
		createCmd.WriteString(" --internal")
	}

	// Add default driver options
	createCmd.WriteString(" --driver bridge")

	// Create network
	cmd = NewRemoteCommand(n.ssh, createCmd.String())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create network %s: %w", name, err)
	}

	return nil
}

func (n *NetworkManager) ConnectContainer(container string, networks []string) error {
	for _, network := range networks {
		// Check if container is already connected
		cmd := NewRemoteCommand(n.ssh, fmt.Sprintf(
			"podman network inspect %s | grep -q %s",
			network, container,
		))
		if err := cmd.Run(); err == nil {
			continue // Already connected
		}

		// Connect container to network
		cmd = NewRemoteCommand(n.ssh, fmt.Sprintf(
			"podman network connect %s %s",
			network, container,
		))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to connect container %s to network %s: %w", container, network, err)
		}
	}

	return nil
}

func (n *NetworkManager) DisconnectContainer(container string, network string) error {
	cmd := NewRemoteCommand(n.ssh, fmt.Sprintf(
		"podman network disconnect %s %s",
		network, container,
	))
	return cmd.Run()
}

func (n *NetworkManager) RemoveNetwork(name string) error {
	cmd := NewRemoteCommand(n.ssh, fmt.Sprintf("podman network rm %s", name))
	return cmd.Run()
}

func (n *NetworkManager) ListNetworks() ([]string, error) {
	cmd := NewRemoteCommand(n.ssh, "podman network ls --format '{{.Name}}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	networks := strings.Split(strings.TrimSpace(string(output)), "\n")
	return networks, nil
}

type Healthcheck struct {
	Test     []string
	Interval string
	Timeout  string
	Retries  int
}

type Container struct {
	Name        string
	Image       string
	Command     []string
	User        string
	Ports       []string
	Volumes     []string
	Environment map[string]string
	EnvFile     []string
	Capabilities []string
	ExtraHosts  []string
	Restart     string
	Healthcheck Healthcheck
}

type VolumeConfig struct {
	Name    string
	Path    string
	Options []string // For :Z etc
}

type ContainerManager struct {
	ssh *SSHClient
}

func NewContainerManager(ssh *SSHClient) *ContainerManager {
	return &ContainerManager{
		ssh: ssh,
	}
}

func (c *ContainerManager) Run(container Container) error {
	cmd := strings.Builder{}
	cmd.WriteString("podman run -d")

	// Name
	cmd.WriteString(fmt.Sprintf(" --name %s", container.Name))

	// User
	if container.User != "" {
		cmd.WriteString(fmt.Sprintf(" --user %s", container.User))
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
		vol := parseVolume(volume)
		if vol.Name != "" {
			createCmd := NewRemoteCommand(c.ssh, fmt.Sprintf("podman volume create %s", vol.Name))
			createCmd.Run()
		}
		
		volStr := vol.Name + ":" + vol.Path
		if len(vol.Options) > 0 {
			var podmanOpts []string
			for _, opt := range vol.Options {
				switch opt {
				case "z", "Z", "ro", "rw",
					 "shared", "slave", "private",
					 "rshared", "rslave", "rprivate":
					podmanOpts = append(podmanOpts, opt)
				}
			}
			if len(podmanOpts) > 0 {
				volStr += ":" + strings.Join(podmanOpts, ",")
			}
		}
		cmd.WriteString(fmt.Sprintf(" -v %s", volStr))
	}

	// Environment
	for _, envFile := range container.EnvFile {
		cmd.WriteString(fmt.Sprintf(" --env-file %s", envFile))
	}
	for key, value := range container.Environment {
		// Expand env vars in values
		value = os.ExpandEnv(value)
		cmd.WriteString(fmt.Sprintf(" -e %s=%s", key, value))
	}

	// Healthcheck
	if len(container.Healthcheck.Test) > 0 {
		cmd.WriteString(fmt.Sprintf(" --health-cmd %q", strings.Join(container.Healthcheck.Test, " ")))
		cmd.WriteString(fmt.Sprintf(" --health-interval %s", container.Healthcheck.Interval))
		cmd.WriteString(fmt.Sprintf(" --health-timeout %s", container.Healthcheck.Timeout))
		cmd.WriteString(fmt.Sprintf(" --health-retries %d", container.Healthcheck.Retries))
	}

	// Image
	cmd.WriteString(fmt.Sprintf(" %s", container.Image))

	// Command
	if len(container.Command) > 0 {
		cmd.WriteString(fmt.Sprintf(" %s", strings.Join(container.Command, " ")))
	}

	execCmd := NewRemoteCommand(c.ssh, cmd.String())
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to be healthy if healthcheck defined
	if len(container.Healthcheck.Test) > 0 {
		return c.waitForHealthy(container.Name)
	}

	return nil
}

func parseVolume(volume string) VolumeConfig {
	parts := strings.Split(volume, ":")
	if len(parts) < 2 {
		return VolumeConfig{}
	}

	var options []string
	if len(parts) > 2 {
		options = strings.Split(parts[2], ",")
	}

	// Check if this is a named volume
	if !strings.Contains(parts[0], "/") {
		return VolumeConfig{
			Name: parts[0],
			Path: parts[1],
			Options: options,
		}
	}

	// Regular path volume
	return VolumeConfig{
		Path: parts[1],
		Options: options,
	}
}

func (c *ContainerManager) waitForHealthy(name string) error {
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		cmd := NewRemoteCommand(c.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Health.Status}}' %s",
			name,
		))
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "healthy" {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("container failed to become healthy")
}

func (c *ContainerManager) Remove(name string) error {
	cmd := NewRemoteCommand(c.ssh, fmt.Sprintf("podman rm -f %s", name))
	return cmd.Run()
}
