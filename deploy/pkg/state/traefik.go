package state

import (
	"github.com/bluetongueai/uberbase/deploy/pkg/containers"
	"github.com/bluetongueai/uberbase/deploy/pkg/traefik"
)

type TraefikState struct {
	Tag     containers.ContainerTag
	Configs map[string]traefik.TraefikDynamicConfiguration
}
