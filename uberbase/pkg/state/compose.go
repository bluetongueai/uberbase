package state

import "github.com/compose-spec/compose-go/v2/types"

type ComposeVolumeState struct {
	Name   string              `yaml:"name"`
	Config *types.VolumeConfig `yaml:"volume"`
}

type ComposeNetworkState struct {
	Name   string               `yaml:"name"`
	Config *types.NetworkConfig `yaml:"network"`
}

type ComposeServiceState struct {
	ServiceName   string `yaml:"service_name"`
	ContainerName string `yaml:"container_name"`
	Hostname      string `yaml:"hostname"`
	Image         string `yaml:"image"`
	Tag           string `yaml:"tag"`
}

type ComposeState struct {
	Services map[string]*ComposeServiceState `yaml:"services"`
	Volumes  map[string]*ComposeVolumeState  `yaml:"volumes"`
	Networks map[string]*ComposeNetworkState `yaml:"networks"`
}
