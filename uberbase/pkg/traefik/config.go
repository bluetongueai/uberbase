package traefik

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	StaticConfigPath  = "/etc/traefik/config/traefik.yml"
	DynamicConfigPath = "/etc/traefik/config/dynamic"
)

type TraefikConfig interface {
	WriteToFile(path string) error
}

func LoadTraefikStaticConfig() (*TraefikStaticConfiguration, error) {
	content, err := os.ReadFile(StaticConfigPath)
	if err != nil {
		return nil, err
	}

	var config TraefikStaticConfiguration
	err = yaml.Unmarshal(content, &config)
	return &config, err
}

func LoadTraefikDynamicConfigs() (map[string]*TraefikDynamicConfiguration, error) {
	configs := make(map[string]*TraefikDynamicConfiguration)

	files, err := os.ReadDir(DynamicConfigPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(DynamicConfigPath, file.Name()))
		if err != nil {
			return nil, err
		}
		var config TraefikDynamicConfiguration
		err = yaml.Unmarshal(content, &config)
		if err != nil {
			return nil, err
		}
		configs[file.Name()] = &config
	}

	return configs, nil
}

func (c *TraefikStaticConfiguration) WriteToFile(path string) error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, content, 0644)
}

func (c *TraefikDynamicConfiguration) WriteToFile(dir, name string) error {
	content, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), content, 0644)
}
