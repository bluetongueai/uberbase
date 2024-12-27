package deploy

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

const (
	BackupLabelPrefix = "bluetongue.backup."
	BackupEnabled     = BackupLabelPrefix + "enabled"
	BackupSchedule    = BackupLabelPrefix + "schedule"
	BackupRetention   = BackupLabelPrefix + "retention"
	BackupType        = BackupLabelPrefix + "type"

	// Host placement labels
	PlacementLabelPrefix = "bluetongue.placement."
	PlacementHost        = PlacementLabelPrefix + "host"  // Specific host to deploy to
	PlacementHosts       = PlacementLabelPrefix + "hosts" // Comma-separated list of hosts
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
	Backup      *BackupConfig                     `yaml:"x-backup,omitempty"`
}

type BackupConfig struct {
	Enabled    bool   `yaml:"enabled"`
	Schedule   string `yaml:"schedule"`
	Retention  string `yaml:"retention"`
	BackupType string `yaml:"type"`
}

type PlacementConfig struct {
	Constraints []PlacementConstraint
}

type PlacementConstraint struct {
	Type      string
	Value     string
	Operation string
}

func extractPlacementConfig(labels map[string]string) *PlacementConfig {
	if labels == nil {
		return &PlacementConfig{}
	}

	config := &PlacementConfig{
		Constraints: make([]PlacementConstraint, 0),
	}

	// Parse single host constraint
	if host := labels[PlacementHost]; host != "" {
		config.Constraints = append(config.Constraints, PlacementConstraint{
			Type:      "host",
			Value:     host,
			Operation: "=",
		})
	}

	// Parse multiple hosts
	if hosts := labels[PlacementHosts]; hosts != "" {
		hostList := strings.Split(hosts, ",")
		for _, host := range hostList {
			config.Constraints = append(config.Constraints, PlacementConstraint{
				Type:      "host",
				Value:     strings.TrimSpace(host),
				Operation: "=",
			})
		}
	}

	return config
}
