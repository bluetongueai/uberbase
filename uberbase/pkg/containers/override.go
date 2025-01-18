package containers

import (
	"fmt"
	"path/filepath"

	"github.com/bluetongueai/uberbase/uberbase/pkg/core"
	"github.com/bluetongueai/uberbase/uberbase/pkg/utils"
	"gopkg.in/yaml.v2"
)

type ComposeServiceOverride struct {
	Hostname string `yaml:"hostname"`
	Image    string `yaml:"image"`
	EnvFile  string `yaml:"env_file"`
	Name     string `yaml:"-"`
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
				Image:    utils.StripTag(service.Image) + ":" + string(containerTag),
				EnvFile:  fmt.Sprintf("%s.env", string(containerTag)),
			}
		}
	}

	return override
}

func (c *ComposeOverride) WriteToFile(executor *core.RemoteExecutor, remoteWorkDir string) (string, error) {
	yaml, err := yaml.Marshal(c)
	if err != nil {
		return "", err
	}
	overrideFile := filepath.Join(remoteWorkDir, "docker-compose.override.yml")
	cmd := "cat <<EOF > " + overrideFile + "\n" + string(yaml) + "\nEOF"
	_, err = executor.Exec(cmd)
	if err != nil {
		return "", err
	}
	return overrideFile, nil
}
