package state

import (
	"github.com/bluetongueai/uberbase/uberbase/pkg/containers"
	"github.com/bluetongueai/uberbase/uberbase/pkg/traefik"
)

type TraefikState struct {
	Tag     containers.ContainerTag
	Configs map[string]traefik.TraefikDynamicConfiguration
}
