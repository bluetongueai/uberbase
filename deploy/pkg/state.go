package pkg

import (
	"fmt"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DeploymentState tracks the state of deployments
type DeploymentState struct {
	LastKnownGood map[string]ServiceState `yaml:"last_known_good"`
}

// ServiceState tracks the state of services
type ServiceState struct {
	Version       string `yaml:"version"`
	BlueVersion   string `yaml:"blue_version"`
	GreenVersion  string `yaml:"green_version"`
	BlueWeight    int    `yaml:"blue_weight"`
	GreenWeight   int    `yaml:"green_weight"`
	LastUpdated   string `yaml:"last_updated"`
}

// StateManager handles saving and loading deployment state
type StateManager struct {
	ssh     SSHClientInterface
	workDir string
}

// NewStateManager creates a new state manager
func NewStateManager(ssh SSHClientInterface, workDir string) *StateManager {
	return &StateManager{
		ssh:     ssh,
		workDir: workDir,
	}
}

// Save persists the deployment state
func (s *StateManager) Save(state DeploymentState) error {
	if err := s.validateState(state); err != nil {
		return fmt.Errorf("state validation failed: %w", err)
	}

	if s.workDir == "" {
		return fmt.Errorf("work directory not set")
	}

	cmd := NewRemoteCommand(s.ssh, fmt.Sprintf("mkdir -p %s", filepath.Dir(s.workDir)))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	stateFile := filepath.Join(s.workDir, "deployment-state.yml")
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := s.ssh.WriteFile(stateFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}
	return nil
}

// Load retrieves the deployment state
func (s *StateManager) Load() (DeploymentState, error) {
	var state DeploymentState
	stateFile := filepath.Join(s.workDir, "deployment-state.yml")
	
	data, err := s.ssh.ReadFile(stateFile)
	if err != nil {
		// Return empty state with initialized map
		return DeploymentState{
			LastKnownGood: make(map[string]ServiceState),
		}, nil
	}

	if err := yaml.Unmarshal(data, &state); err != nil {
		return DeploymentState{}, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	// Initialize map if nil
	if state.LastKnownGood == nil {
		state.LastKnownGood = make(map[string]ServiceState)
	}
	
	return state, nil
}

func (s *StateManager) validateState(state DeploymentState) error {
	if state.LastKnownGood == nil {
		return nil // Allow empty state
	}

	for service, serviceState := range state.LastKnownGood {
		// Only validate weights if both are non-zero (active blue/green deployment)
		if serviceState.BlueWeight != 0 || serviceState.GreenWeight != 0 {
			if serviceState.BlueWeight < 0 || serviceState.BlueWeight > 100 ||
			   serviceState.GreenWeight < 0 || serviceState.GreenWeight > 100 {
				return fmt.Errorf("invalid weight for service %s", service)
			}
			
			if serviceState.BlueWeight+serviceState.GreenWeight != 100 {
				return fmt.Errorf("weights must sum to 100 for service %s", service)
			}
		}
		
		// Version can be empty for new services
		if serviceState.Version == "" && (serviceState.BlueVersion != "" || serviceState.GreenVersion != "") {
			return fmt.Errorf("invalid version for service %s", service)
		}
	}
	return nil
}
