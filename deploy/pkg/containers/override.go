package containers

import (
	"github.com/bluetongueai/uberbase/deploy/pkg"
	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"gopkg.in/yaml.v2"
)

type ComposeServiceOverride struct {
	Name     string `yaml:"name"`
	Hostname string `yaml:"hostname"`
	Image    string `yaml:"image"`
	RefName  string `yaml:"-"`
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
				Image:    pkg.StripTag(service.Image) + ":" + string(containerTag),
			}
		}
	}

	return override
}

func (c *ComposeOverride) WriteToFile(executor *core.RemoteExecutor, remoteWorkDir, filePath string) (string, error) {
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	overrideFile := remoteWorkDir + "/docker-compose.override.yml"
	cmd := "cat <<EOF > " + overrideFile + "\n" + string(yaml) + "\nEOF"
	_, err = executor.Exec(cmd)
	if err != nil {
		return "", err
	}
	return overrideFile, nil
}
