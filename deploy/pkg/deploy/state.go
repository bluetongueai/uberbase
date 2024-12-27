package deploy

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

// DeploymentState tracks the state of deployments
type DeploymentState struct {
	Services map[string]*ServiceState `yaml:"services"`
}

// StateManager handles saving and loading deployment state
type StateManager struct {
	ssh     *core.SSHConnection
	workDir string
}

// NewStateManager creates a new state manager
func NewStateManager(ssh *core.SSHConnection, workDir string) *StateManager {
	core.Logger.Debug("Creating new StateManager")
	return &StateManager{
		ssh:     ssh,
		workDir: workDir,
	}
}

// Save persists the deployment state
func (s *StateManager) Save(state DeploymentState) error {
	core.Logger.Info("Saving deployment state")
	if err := s.validateState(state); err != nil {
		core.Logger.Errorf("State validation failed: %v", err)
		return fmt.Errorf("state validation failed: %w", err)
	}

	if s.workDir == "" {
		core.Logger.Error("Work directory not set")
		return fmt.Errorf("work directory not set")
	}

	_, err := s.ssh.Exec(fmt.Sprintf("mkdir -p %s", filepath.Dir(s.workDir)))
	if err != nil {
		core.Logger.Errorf("Failed to create state directory: %v", err)
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	stateFile := filepath.Join(s.workDir, "deployment-state.yml")
	data, err := yaml.Marshal(state)
	if err != nil {
		core.Logger.Errorf("Failed to marshal state: %v", err)
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	result, err := s.ssh.Exec(fmt.Sprintf("echo '%s' > %s", string(data), stateFile))
	if err != nil {
		core.Logger.Errorf("Failed to write state file: %v", err)
		return fmt.Errorf("failed to write state file: %w", err)
	}

	core.Logger.Info("State file written successfully: ", result)
	return nil
}

// Load retrieves the deployment state
func (s *StateManager) Load() (DeploymentState, error) {
	core.Logger.Info("Loading deployment state")
	var state DeploymentState
	stateFile := filepath.Join(s.workDir, "deployment-state.yml")

	data, err := s.ssh.Exec(fmt.Sprintf("cat %s", stateFile))
	if err != nil {
		core.Logger.Warn("State file not found, returning empty state")
		return DeploymentState{
			Services: make(map[string]*ServiceState),
		}, nil
	}

	if err := yaml.Unmarshal([]byte(data), &state); err != nil {
		core.Logger.Errorf("Failed to unmarshal state: %v", err)
		return DeploymentState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Initialize map if nil
	if state.Services == nil {
		state.Services = make(map[string]*ServiceState)
	}

	return state, nil
}

func (s *StateManager) validateState(state DeploymentState) error {
	if state.Services == nil {
		return nil // Allow empty state
	}

	for serviceName, state := range state.Services {
		// Only validate weights if both are non-zero (active blue/green deployment)
		if state.BlueWeight != 0 || state.GreenWeight != 0 {
			if state.BlueWeight < 0 || state.BlueWeight > 100 ||
				state.GreenWeight < 0 || state.GreenWeight > 100 {
				return fmt.Errorf("invalid weight for service %s", serviceName)
			}

			if state.BlueWeight+state.GreenWeight != 100 {
				return fmt.Errorf("weights must sum to 100 for service %s", serviceName)
			}
		}
	}
	return nil
}
