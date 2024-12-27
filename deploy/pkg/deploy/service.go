package deploy

import "github.com/bluetongueai/uberbase/deploy/pkg/podman"

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
}

// ServiceState moved from state.go
type ServiceState struct {
	BlueVersion  string `yaml:"blue_version"`
	GreenVersion string `yaml:"green_version"`
	BlueWeight   int    `yaml:"blue_weight"`
	GreenWeight  int    `yaml:"green_weight"`
	LastUpdated  string `yaml:"last_updated"`
}

// ConvertComposeService converts a ComposeService to a Service
func ConvertComposeService(name string, cs ComposeService) (*Service, error) {
	service := &Service{
		Name:  name,
		Image: podman.ParseImageRef(cs.Image),
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
