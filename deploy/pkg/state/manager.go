package state

import (
	"fmt"
	"path/filepath"
	"reflect"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bluetongueai/uberbase/deploy/pkg/containers"
	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	"github.com/bluetongueai/uberbase/deploy/pkg/traefik"
)

type DeploymentState struct {
	Tag     containers.ContainerTag `yaml:"tag"`
	Compose *ComposeState           `yaml:"compose"`
	Traefik *TraefikState           `yaml:"traefik"`
	Lock    *DeploymentLock         `yaml:"lock,omitempty"`
}

type DeploymentLock struct {
	AcquiredAt string `yaml:"acquired_at"`
	ExpiresAt  string `yaml:"expires_at"`
	Owner      string `yaml:"owner"`
	Renewable  bool   `yaml:"renewable"`
}

type StateManager struct {
	CurrentState DeploymentState
	workDir      string
	executor     core.Executor
}

func NewStateManager(workDir string, executor core.Executor) *StateManager {
	logging.Logger.Debug("Creating new StateManager")
	manager := &StateManager{
		workDir:  workDir,
		executor: executor,
	}
	manager.Load()
	return manager
}

func (s *StateManager) Load() (DeploymentState, error) {
	logging.Logger.Info("Loading existing deployment state")
	var state DeploymentState
	stateFile := filepath.Join(s.workDir, "deployment-state.yml")

	data, err := s.executor.Exec(fmt.Sprintf("cat %s", stateFile))
	if err != nil {
		logging.Logger.Info("State file not found, initializing empty state")
		return DeploymentState{
			Compose: &ComposeState{
				Services: make(map[string]*ComposeServiceState),
			},
			Traefik: &TraefikState{
				Tag: "",
			},
		}, nil
	}

	if err := yaml.Unmarshal([]byte(data), &state); err != nil {
		logging.Logger.Errorf("Failed to unmarshal state: %v", err)
		return DeploymentState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return state, nil
}

func (s *StateManager) Update(services map[string]containers.ComposeServiceOverride, dynamicConfigs map[string]*traefik.TraefikDynamicConfiguration, tag containers.ContainerTag) error {
	// update the state's compose config with the new override services
	for _, service := range s.CurrentState.Compose.Services {
		if override, ok := services[service.ServiceName]; ok {
			service.ServiceName = override.RefName + "-" + string(tag)
			service.ContainerName = override.RefName + "-" + string(tag)
			service.Image = override.Image
			service.Hostname = override.Hostname
		}
	}

	// update the state's traefik config with the new dynamic configs
	s.CurrentState.Traefik.Tag = tag
	s.CurrentState.Traefik.Configs = make(map[string]traefik.TraefikDynamicConfiguration)
	for name, config := range dynamicConfigs {
		s.CurrentState.Traefik.Configs[name] = *config
	}

	return s.Save()
}

func (s *StateManager) Save() error {
	return s.write(s.CurrentState)
}

func (s *StateManager) write(state DeploymentState) error {
	if s.workDir == "" {
		logging.Logger.Error("Work directory not set")
		return fmt.Errorf("work directory not set")
	}

	_, err := s.executor.Exec(fmt.Sprintf("mkdir -p %s", filepath.Dir(s.workDir)))
	if err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	logging.Logger.Info("Saving deployment state")

	// Marshal state to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first
	tempFile := filepath.Join(s.workDir, fmt.Sprintf("deployment-state.yml.%d.tmp", time.Now().UnixNano()))
	finalFile := filepath.Join(s.workDir, "deployment-state.yml")

	// Write data to temp file
	if _, err := s.executor.Exec(fmt.Sprintf("echo '%s' > %s", string(data), tempFile)); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Atomically rename temp file to final file
	if _, err := s.executor.Exec(fmt.Sprintf("mv -f %s %s", tempFile, finalFile)); err != nil {
		// Try to cleanup temp file
		s.executor.Exec(fmt.Sprintf("rm -f %s", tempFile))
		return fmt.Errorf("failed to atomically update state file: %w", err)
	}

	logging.Logger.Info("State file updated successfully")
	return nil
}

// Equal compares two states for equality
func (s *DeploymentState) Equal(other *DeploymentState) bool {
	if s.Tag != other.Tag {
		return false
	}

	if len(s.Compose.Services) != len(other.Compose.Services) {
		return false
	}

	// Compare services
	for name, service := range s.Compose.Services {
		otherService, exists := other.Compose.Services[name]
		if !exists || service.ContainerName != otherService.ContainerName {
			return false
		}
	}

	// Compare Traefik configs
	return reflect.DeepEqual(s.Traefik, other.Traefik)
}
