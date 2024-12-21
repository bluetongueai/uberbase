package pkg

import (
	"fmt"
	"strings"
)

type NetworkManager struct {
	ssh SSHClientInterface
}

func NewNetworkManager(ssh SSHClientInterface) *NetworkManager {
	return &NetworkManager{
		ssh: ssh,
	}
}

func (n *NetworkManager) validateNetworkName(name string) error {
	if name == "" {
		return fmt.Errorf("network name cannot be empty")
	}
	if len(name) > 63 {
		return fmt.Errorf("network name too long")
	}
	// Check for invalid characters
	if strings.ContainsAny(name, "@$#!%^&*()+=<>{}[]\\|;:'\",/? ") {
		return fmt.Errorf("invalid network name: contains invalid characters")
	}
	return nil
}

func (n *NetworkManager) EnsureNetwork(name string, internal bool) error {
	if err := n.validateNetworkName(name); err != nil {
		return err
	}

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
		err := cmd.Run()
		if err == nil {
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
