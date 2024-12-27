package deploy

import (
	"os"

	"gopkg.in/yaml.v3"
)

type HostGroupConfig struct {
	HostGroups map[string]HostGroup `yaml:"host_groups"`
}

type HostGroup struct {
	Hosts      []string          `yaml:"hosts"`
	Attributes map[string]string `yaml:"attributes"`
}

func LoadConfig(path string) (*HostGroupConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config HostGroupConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
