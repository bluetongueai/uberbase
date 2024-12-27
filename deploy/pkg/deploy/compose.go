package deploy

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

type ComposeConfig struct {
	Version  string                    `yaml:"version,omitempty"`
	Services map[string]ComposeService `yaml:"services"`
	Volumes  map[string]ComposeVolume  `yaml:"volumes,omitempty"`
	Networks map[string]ComposeNetwork `yaml:"networks,omitempty"`
}

type ComposeVolume struct {
	Driver string `yaml:"driver,omitempty"`
}

type ComposeNetwork struct {
	Driver   string `yaml:"driver,omitempty"`
	Internal bool   `yaml:"internal,omitempty"`
}

type BuildConfig struct {
	Context    string            `yaml:"context"`
	Dockerfile string            `yaml:"dockerfile,omitempty"`
	Args       map[string]string `yaml:"args,omitempty"`
}

type DependsOnConfig struct {
	Condition string `yaml:"condition"`
	Service   string `yaml:"service"`
}

func ParseComposeFile(path string) (*ComposeConfig, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var composeConfig ComposeConfig
	err = yaml.Unmarshal(yamlFile, &composeConfig)
	return &composeConfig, err
}

type ComposeService struct {
	Build       *BuildConfig                      `yaml:"build,omitempty"`
	Image       string                            `yaml:"image"`
	Command     interface{}                       `yaml:"command,omitempty"`
	User        string                            `yaml:"user,omitempty"`
	Ports       []string                          `yaml:"ports,omitempty"`
	Volumes     []string                          `yaml:"volumes,omitempty"`
	Environment interface{}                       `yaml:"environment,omitempty"`
	EnvFile     interface{}                       `yaml:"env_file,omitempty"`
	DependsOn   map[string]podman.DependsOnConfig `yaml:"depends_on,omitempty"`
	CapAdd      []string                          `yaml:"cap_add,omitempty"`
	ExtraHosts  []string                          `yaml:"extra_hosts,omitempty"`
	Restart     string                            `yaml:"restart,omitempty"`
	Healthcheck *podman.Healthcheck               `yaml:"healthcheck,omitempty"`
	Networks    []string                          `yaml:"networks,omitempty"`
	Domains     []string                          `yaml:"domains,omitempty"`
	SSL         bool                              `yaml:"ssl,omitempty"`
	Private     bool                              `yaml:"private,omitempty"`
	Labels      map[string]string                 `yaml:"labels,omitempty"`
	Resources   ResourceConfig                    `yaml:"resources,omitempty"`
	Pod         string                            `yaml:"x-pod,omitempty"`
	SharePID    bool                              `yaml:"x-share-pid,omitempty"`
	IsInit      bool                              `yaml:"x-init,omitempty"`
}
