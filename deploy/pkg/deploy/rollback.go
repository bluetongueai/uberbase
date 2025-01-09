package deploy

import (
	"context"
	"fmt"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	"github.com/bluetongueai/uberbase/deploy/pkg/state"
)

// RollbackFunc represents a function that performs a rollback operation
type RollbackFunc func(context.Context) error

// RollbackStep represents a single rollback step
type RollbackStep struct {
	name     string
	rollback RollbackFunc
	verify   func(context.Context) error // Add verification step
}

// RollbackManager handles the registration and execution of rollback functions
type RollbackManager struct {
	steps []RollbackStep
	state *state.StateManager // Add state manager reference
}

// NewRollbackManager creates a new RollbackManager
func NewRollbackManager(state *state.StateManager) *RollbackManager {
	return &RollbackManager{
		steps: make([]RollbackStep, 0),
		state: state,
	}
}

// AddRollbackStep registers a new rollback function
func (rm *RollbackManager) AddRollbackStep(name string, fn RollbackFunc, verify func(context.Context) error) {
	rm.steps = append(rm.steps, RollbackStep{
		name:     name,
		rollback: fn,
		verify:   verify,
	})
}

// Rollback executes all registered rollback functions in reverse order
func (rm *RollbackManager) Rollback(ctx context.Context) error {
	logging.Logger.Info("Rolling back deployment")
	var errors []error

	initialState, _ := rm.state.Load()

	// Execute rollbacks in reverse order
	for i := len(rm.steps) - 1; i >= 0; i-- {
		step := rm.steps[i]

		// Execute rollback with timeout
		timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
		defer cancel()

		if err := step.rollback(timeoutCtx); err != nil {
			errors = append(errors, fmt.Errorf("rollback step '%s' failed: %w", step.name, err))
			continue
		}

		// Verify the rollback step
		if step.verify != nil {
			if err := step.verify(timeoutCtx); err != nil {
				errors = append(errors, fmt.Errorf("rollback verification '%s' failed: %w", step.name, err))
			}
		}
	}

	// Verify final state matches initial state
	finalState, err := rm.state.Load()
	if err != nil {
		errors = append(errors, fmt.Errorf("failed to load final state: %w", err))
	} else if !initialState.Equal(&finalState) {
		errors = append(errors, fmt.Errorf("final state does not match initial state"))
	}

	if len(errors) > 0 {
		return fmt.Errorf("rollback failed with %d errors: %v", len(errors), errors)
	}
	return nil
}
