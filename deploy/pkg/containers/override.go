package containers

import (
	"os"

	"gopkg.in/yaml.v2"
)

type ComposeServiceOverride struct {
	RefName  string `yaml:"ref_name"`
	Name     string `yaml:"name"`
	Hostname string `yaml:"hostname"`
	Image    string `yaml:"image"`
}

type ComposeOverride struct {
	Services map[string]ComposeServiceOverride `yaml:"services"`
}

func NewComposeOverride(compose *ComposeProject, containerTag ContainerTag) *ComposeOverride {
	override := &ComposeOverride{
		Services: make(map[string]ComposeServiceOverride),
	}

	for _, service := range compose.Project.Services {
		if service.Build != nil {
			override.Services[service.Name] = ComposeServiceOverride{
				RefName:  service.Name,
				Name:     service.Name + "-" + string(containerTag),
				Hostname: service.Hostname + "-" + string(containerTag),
				Image:    service.Image + ":" + string(containerTag),
			}
		}
	}

	return override
}

func (c *ComposeOverride) WriteToFile(filePath string) (string, error) {
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	err = os.WriteFile(filePath, yaml, 0644)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
