package deploy

import (
	"fmt"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

// DeploymentState tracks the state of deployments
type DeploymentState struct {
	Services map[string]*ServiceState `yaml:"services"`
	Lock     *DeploymentLock          `yaml:"lock,omitempty"`
}

type DeploymentLock struct {
	AcquiredAt string `yaml:"acquired_at"`
	ExpiresAt  string `yaml:"expires_at"`
	Owner      string `yaml:"owner"`
	Renewable  bool   `yaml:"renewable"`
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

	// Ensure directory exists
	_, err := s.ssh.Exec(fmt.Sprintf("mkdir -p %s", filepath.Dir(s.workDir)))
	if err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Marshal state to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to temporary file first
	tempFile := filepath.Join(s.workDir, fmt.Sprintf("deployment-state.yml.%d.tmp", time.Now().UnixNano()))
	finalFile := filepath.Join(s.workDir, "deployment-state.yml")

	// Write data to temp file
	if _, err := s.ssh.Exec(fmt.Sprintf("echo '%s' > %s", string(data), tempFile)); err != nil {
		return fmt.Errorf("failed to write temporary state file: %w", err)
	}

	// Atomically rename temp file to final file
	if _, err := s.ssh.Exec(fmt.Sprintf("mv -f %s %s", tempFile, finalFile)); err != nil {
		// Try to cleanup temp file
		s.ssh.Exec(fmt.Sprintf("rm -f %s", tempFile))
		return fmt.Errorf("failed to atomically update state file: %w", err)
	}

	core.Logger.Info("State file updated successfully")
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

func (s *StateManager) AcquireLock(owner string) error {
	state, err := s.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Check if lock exists and is still valid
	if state.Lock != nil {
		expiresAt, err := time.Parse(time.RFC3339, state.Lock.ExpiresAt)
		if err == nil && time.Now().Before(expiresAt) {
			return fmt.Errorf("deployment in progress by %s until %s",
				state.Lock.Owner, state.Lock.ExpiresAt)
		}
	}

	// Acquire lock with 1 hour timeout
	state.Lock = &DeploymentLock{
		AcquiredAt: time.Now().UTC().Format(time.RFC3339),
		ExpiresAt:  time.Now().Add(1 * time.Hour).UTC().Format(time.RFC3339),
		Owner:      owner,
		Renewable:  true,
	}

	return s.Save(state)
}

func (s *StateManager) ReleaseLock(owner string) error {
	state, err := s.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.Lock != nil && state.Lock.Owner == owner {
		state.Lock = nil
		return s.Save(state)
	}

	return fmt.Errorf("lock not held by %s", owner)
}

func (s *StateManager) ExtendLock(owner string, duration time.Duration) error {
	state, err := s.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	if state.Lock == nil || state.Lock.Owner != owner {
		return fmt.Errorf("lock not held by %s", owner)
	}

	if !state.Lock.Renewable {
		return fmt.Errorf("lock is not renewable")
	}

	// Extend lock
	state.Lock.ExpiresAt = time.Now().Add(duration).UTC().Format(time.RFC3339)
	return s.Save(state)
}

// Add auto-extension capability
type LockKeepAlive struct {
	owner        string
	interval     time.Duration
	stopChan     chan struct{}
	stateManager *StateManager
}

func NewLockKeepAlive(owner string, stateManager *StateManager) *LockKeepAlive {
	return &LockKeepAlive{
		owner:        owner,
		interval:     15 * time.Minute,
		stopChan:     make(chan struct{}),
		stateManager: stateManager,
	}
}

func (l *LockKeepAlive) Start() {
	go func() {
		ticker := time.NewTicker(l.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := l.stateManager.ExtendLock(l.owner, time.Hour); err != nil {
					core.Logger.Errorf("Failed to extend lock: %v", err)
				}
			case <-l.stopChan:
				return
			}
		}
	}()
}

func (l *LockKeepAlive) Stop() {
	close(l.stopChan)
}
