package podman

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type NetworkManager struct {
	ssh *core.SSHConnection
}

type Network struct {
	Name     string
	Internal bool
}

func NewNetworkManager(ssh *core.SSHConnection) *NetworkManager {
	core.Logger.Debug("Creating new NetworkManager")
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
	core.Logger.Infof("Ensuring network: %s", name)
	if err := n.validateNetworkName(name); err != nil {
		core.Logger.Errorf("Invalid network name: %v", err)
		return err
	}

	// Check if network exists
	_, err := n.ssh.Exec(fmt.Sprintf("podman network ls | grep -q %s", name))
	if err == nil {
		core.Logger.Infof("Network already exists: %s", name)
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
	_, err = n.ssh.Exec(createCmd.String())
	if err != nil {
		core.Logger.Errorf("Failed to create network %s: %v", name, err)
		return fmt.Errorf("failed to create network %s: %w", name, err)
	}

	core.Logger.Infof("Network created successfully: %s", name)
	return nil
}

func (n *NetworkManager) ConnectContainer(container string, networks []string) error {
	core.Logger.Infof("Connecting container %s to networks: %v", container, networks)
	for _, network := range networks {
		// Check if container is already connected
		_, err := n.ssh.Exec(fmt.Sprintf(
			"podman network inspect %s | grep -q %s",
			network, container,
		))
		if err == nil {
			core.Logger.Infof("Container %s already connected to network %s", container, network)
			continue // Already connected
		}

		// Connect container to network
		_, err = n.ssh.Exec(fmt.Sprintf(
			"podman network connect %s %s",
			network, container,
		))
		if err != nil {
			core.Logger.Errorf("Failed to connect container %s to network %s: %v", container, network, err)
			return fmt.Errorf("failed to connect container %s to network %s: %w", container, network, err)
		}
	}

	core.Logger.Infof("Container %s connected to networks successfully", container)
	return nil
}

func (n *NetworkManager) DisconnectContainer(container string, network string) error {
	core.Logger.Infof("Disconnecting container %s from network %s", container, network)
	_, err := n.ssh.Exec(fmt.Sprintf(
		"podman network disconnect %s %s",
		network, container,
	))
	if err != nil {
		core.Logger.Errorf("Failed to disconnect container %s from network %s: %v", container, network, err)
		return err
	}
	return nil
}

func (n *NetworkManager) RemoveNetwork(name string) error {
	core.Logger.Infof("Removing network: %s", name)
	_, err := n.ssh.Exec(fmt.Sprintf("podman network rm %s", name))
	if err != nil {
		core.Logger.Errorf("Failed to remove network %s: %v", name, err)
		return err
	}
	return nil
}

func (n *NetworkManager) ListNetworks() ([]string, error) {
	core.Logger.Info("Listing networks")
	output, err := n.ssh.Exec("podman network ls --format '{{.Name}}'")
	if err != nil {
		core.Logger.Errorf("Failed to list networks: %v", err)
		return nil, err
	}

	networks := strings.Split(strings.TrimSpace(string(output)), "\n")
	return networks, nil
}
