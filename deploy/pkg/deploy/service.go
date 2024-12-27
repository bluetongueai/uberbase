package deploy

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

/*
A service is a container with volumes in a network that is deployed into the proxy
with a configuration.
*/

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
	Count        *int
	Device       string
	Driver       string
	Options      map[string]string
}

type Service struct {
	Name         string
	Image        podman.ImageRef
	Pod          *podman.Pod
	Container    *podman.Container
	Environment  *Environment
	Volumes      []*podman.Volume
	Network      *podman.Network
	DeployConfig *DeployConfig
	State        *ServiceState
	DependsOn    map[string]podman.DependsOnConfig
	Placement    *PlacementConfig
}

// ServiceState moved from state.go
type ServiceState struct {
	BlueVersion  string                  `yaml:"blue_version"`
	GreenVersion string                  `yaml:"green_version"`
	BlueWeight   int                     `yaml:"blue_weight"`
	GreenWeight  int                     `yaml:"green_weight"`
	LastUpdated  string                  `yaml:"last_updated"`
	VolumeStates map[string]*VolumeState `yaml:"volume_states,omitempty"`
	RollbackInfo *RollbackInfo           `yaml:"rollback_info,omitempty"`
	Transactions []TransactionLog        `yaml:"transactions,omitempty"`
}

// Add VolumeState struct
type VolumeState struct {
	Name       string `yaml:"name"`
	BackupPath string `yaml:"backup_path,omitempty"`
	LastBackup string `yaml:"last_backup,omitempty"`
}

type RollbackInfo struct {
	PreviousVersion string            `yaml:"previous_version"`
	VolumeBackups   map[string]string `yaml:"volume_backups,omitempty"`
	Timestamp       string            `yaml:"timestamp"`
}

// ConvertComposeService converts a ComposeService to a Service
func ConvertComposeService(name string, cs ComposeService) (*Service, error) {
	service := &Service{
		Name:      name,
		Image:     podman.ParseImageRef(cs.Image),
		Placement: extractPlacementConfig(cs.Labels),
		Environment: &Environment{
			Variables: parseEnvironment(cs.Environment),
			Files:     parseEnvFile(cs.EnvFile),
		},
		State: &ServiceState{},
	}

	// Convert container config
	command, _ := cs.Command.([]string) // Handle nil case
	container := &podman.Container{
		Name:         name,
		Image:        cs.Image,
		Command:      command,
		User:         cs.User,
		Ports:        cs.Ports,
		Volumes:      cs.Volumes,
		Environment:  parseEnvironment(cs.Environment),
		EnvFile:      parseEnvFile(cs.EnvFile),
		Capabilities: cs.CapAdd,
		ExtraHosts:   cs.ExtraHosts,
		Restart:      cs.Restart,
		Healthcheck:  cs.Healthcheck, // Use existing Healthcheck from compose
		Networks:     cs.Networks,
		Labels:       cs.Labels,
		DependsOn:    cs.DependsOn,
	}
	service.Container = container
	service.DependsOn = cs.DependsOn

	// Handle pod if specified
	if cs.Pod != "" {
		service.Pod = &podman.Pod{
			Name:     cs.Pod,
			Networks: cs.Networks,
			Labels:   cs.Labels,
			Ports:    cs.Ports,
			// Pod-specific settings from compose extensions
			SharePID: cs.SharePID,
			Infra:    true, // Always create infra container
			// Resource limits from compose
			Resources: podman.ResourceLimits{
				CPUQuota:  cs.Resources.Limits.CPUs,
				Memory:    cs.Resources.Limits.Memory,
				CPUShares: cs.Resources.Reservations.CPUs,
			},
		}
	}

	return service, nil
}

// Add these validation functions
func (s *Service) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if s.Container == nil {
		return fmt.Errorf("container configuration is required")
	}

	// Network validation
	if s.Network != nil {
		if err := validateNetworkConfig(s.Network); err != nil {
			return fmt.Errorf("invalid network configuration: %w", err)
		}
	}

	// Volume validation
	for _, vol := range s.Volumes {
		if err := validateVolumeConfig(vol); err != nil {
			return fmt.Errorf("invalid volume configuration: %w", err)
		}
	}

	// Validate dependencies
	if err := s.validateDependencies(); err != nil {
		return fmt.Errorf("invalid dependencies: %w", err)
	}

	return nil
}

func validateNetworkConfig(network *podman.Network) error {
	if network.Name == "" {
		return fmt.Errorf("network name is required")
	}
	if strings.ContainsAny(network.Name, " $&\"'") {
		return fmt.Errorf("network name contains invalid characters")
	}
	return nil
}

func validateVolumeConfig(volume *podman.Volume) error {
	if volume.Name == "" {
		return fmt.Errorf("volume name is required")
	}
	if strings.ContainsAny(volume.Name, " $&\"'") {
		return fmt.Errorf("volume name contains invalid characters")
	}
	return nil
}

func (s *Service) validateDependencies() error {
	if s.DependsOn == nil {
		return nil
	}

	for dep, config := range s.DependsOn {
		if dep == s.Name {
			return fmt.Errorf("service cannot depend on itself")
		}
		if config.Condition != "service_started" &&
			config.Condition != "service_healthy" &&
			config.Condition != "service_completed_successfully" {
			return fmt.Errorf("invalid dependency condition: %s", config.Condition)
		}
	}
	return nil
}
