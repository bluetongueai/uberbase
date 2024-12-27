package podman

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type Pod struct {
	Name       string
	Hostname   string
	Networks   []string
	DNSServers []string
	DNSSearch  []string
	Labels     map[string]string
	Ports      []string

	// Add pod-specific features
	SharePID  bool           // Share PID namespace between containers
	Infra     bool           // Create infra container (default true)
	InitImage string         // Custom init container image
	Resources ResourceLimits // Pod-level resource limits
}

type ResourceLimits struct {
	CPUQuota  string
	Memory    string
	CPUShares string
}

type PodManager struct {
	ssh            *core.SSHConnection
	networkManager *NetworkManager
}

func NewPodManager(ssh *core.SSHConnection) *PodManager {
	core.Logger.Debug("Creating new PodManager")
	return &PodManager{
		ssh:            ssh,
		networkManager: NewNetworkManager(ssh),
	}
}

func (p *PodManager) Create(pod Pod) error {
	core.Logger.Infof("Creating pod: %s", pod.Name)
	if err := p.validatePod(pod); err != nil {
		core.Logger.Errorf("Invalid pod configuration: %v", err)
		return fmt.Errorf("invalid pod configuration: %w", err)
	}

	cmd := strings.Builder{}
	cmd.WriteString(fmt.Sprintf("podman pod create --name %s", pod.Name))

	// Basic configurations
	if pod.Hostname != "" {
		cmd.WriteString(fmt.Sprintf(" --hostname %s", pod.Hostname))
	}

	// Network configurations
	if len(pod.Networks) > 0 {
		cmd.WriteString(fmt.Sprintf(" --network %s", pod.Networks[0]))
	}

	// DNS configurations
	for _, dns := range pod.DNSServers {
		cmd.WriteString(fmt.Sprintf(" --dns %s", dns))
	}
	for _, search := range pod.DNSSearch {
		cmd.WriteString(fmt.Sprintf(" --dns-search %s", search))
	}

	// Labels
	for k, v := range pod.Labels {
		cmd.WriteString(fmt.Sprintf(" --label %s=%s", k, v))
	}

	// Ports
	for _, port := range pod.Ports {
		cmd.WriteString(fmt.Sprintf(" -p %s", port))
	}

	// Execute the command
	_, err := p.ssh.Exec(cmd.String())
	if err != nil {
		core.Logger.Errorf("Failed to create pod: %v", err)
		// Cleanup on failure
		_, _ = p.ssh.Exec(fmt.Sprintf("podman pod rm -f %s", pod.Name))
		return fmt.Errorf("failed to create pod: %w", err)
	}

	// Handle additional networks after pod is created
	if len(pod.Networks) > 1 {
		for _, network := range pod.Networks[1:] {
			_, err := p.ssh.Exec(fmt.Sprintf(
				"podman network connect %s %s",
				network, pod.Name,
			))
			if err != nil {
				core.Logger.Errorf("Failed to connect network %s: %v", network, err)
				return fmt.Errorf("failed to connect network %s: %w", network, err)
			}
		}
	}

	core.Logger.Infof("Pod %s created successfully", pod.Name)
	return nil
}

func (p *PodManager) Remove(name string) error {
	core.Logger.Infof("Removing pod: %s", name)
	_, err := p.ssh.Exec(fmt.Sprintf("podman pod rm -f %s", name))
	if err != nil {
		core.Logger.Errorf("Failed to remove pod %s: %v", name, err)
		return fmt.Errorf("failed to remove pod: %w", err)
	}
	return nil
}

func (p *PodManager) Exists(name string) (bool, error) {
	core.Logger.Debugf("Checking if pod exists: %s", name)
	output, err := p.ssh.Exec(fmt.Sprintf("podman pod exists %s", name))
	if err != nil {
		if strings.Contains(string(output), "no such pod") {
			return false, nil
		}
		core.Logger.Errorf("Error checking pod existence: %v", err)
		return false, fmt.Errorf("error checking pod existence: %w", err)
	}
	return true, nil
}

func (p *PodManager) List() ([]string, error) {
	core.Logger.Debug("Listing pods")
	output, err := p.ssh.Exec("podman pod ls --format '{{.Name}}'")
	if err != nil {
		core.Logger.Errorf("Failed to list pods: %v", err)
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	pods := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(pods) == 1 && pods[0] == "" {
		return []string{}, nil
	}
	return pods, nil
}

func (p *PodManager) validatePod(pod Pod) error {
	if pod.Name == "" {
		return fmt.Errorf("pod name is required")
	}

	return nil
}
