package pkg

import (
	"errors"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestNewStateManager(t *testing.T) {
	ssh := NewMockSSHClient()
	workDir := "/tmp/test"
	
	sm := NewStateManager(ssh, workDir)
	
	assert.NotNil(t, sm)
	assert.Equal(t, ssh, sm.ssh)
	assert.Equal(t, workDir, sm.workDir)
}

func TestSave(t *testing.T) {
	tests := []struct {
		name          string
		state         DeploymentState
		mkdirError    error
		writeError    error
		expectedError bool
	}{
		{
			name: "successful save",
			state: DeploymentState{
				LastKnownGood: map[string]ServiceState{
					"service1": {
						Version:     "nginx:latest",
						BlueWeight:  100,
						GreenWeight: 0,
					},
				},
			},
			mkdirError:    nil,
			writeError:    nil,
			expectedError: false,
		},
		{
			name: "mkdir fails",
			state: DeploymentState{
				LastKnownGood: make(map[string]ServiceState),
			},
			mkdirError:    errors.New("mkdir failed"),
			writeError:    nil,
			expectedError: true,
		},
		{
			name: "write fails",
			state: DeploymentState{
				LastKnownGood: make(map[string]ServiceState),
			},
			mkdirError:    nil,
			writeError:    errors.New("write failed"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssh := NewMockSSHClient()
			workDir := "/tmp/test"
			stateFile := filepath.Join(workDir, "deployment-state.yml")
			
			// Setup mock expectations
			mkdirCmd := fmt.Sprintf("mkdir -p %s", filepath.Dir(workDir))
			if tt.mkdirError != nil {
				ssh.SetError(mkdirCmd, tt.mkdirError)
			}
			
			// Mock WriteFile by setting up RunCommand response
			if tt.writeError != nil {
				ssh.SetError("write "+stateFile, tt.writeError)
			}
			
			sm := NewStateManager(ssh, workDir)
			err := sm.Save(tt.state)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			// Verify mkdir command was called
			commands := ssh.GetCommands()
			assert.Contains(t, commands, mkdirCmd)
		})
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name          string
		fileExists    bool
		fileContent   string
		readError     error
		expectedState DeploymentState
		expectedError bool
	}{
		{
			name:       "successful load",
			fileExists: true,
			fileContent: `last_known_good:
  service1:
    version: "nginx:latest"
    blue_weight: 100
    green_weight: 0
    blue_version: ""
    green_version: ""
    last_updated: ""`,
			readError: nil,
			expectedState: DeploymentState{
				LastKnownGood: map[string]ServiceState{
					"service1": {
						Version:     "nginx:latest",
						BlueWeight:  100,
						GreenWeight: 0,
					},
				},
			},
			expectedError: false,
		},
		{
			name:       "file doesn't exist",
			fileExists: false,
			readError:  errors.New("file not found"),
			expectedState: DeploymentState{
				LastKnownGood: make(map[string]ServiceState),
			},
			expectedError: false,
		},
		{
			name:        "invalid yaml",
			fileExists:  true,
			fileContent: "invalid: yaml: content",
			readError:   nil,
			expectedState: DeploymentState{
				LastKnownGood: nil,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ssh := NewMockSSHClient()
			workDir := "/tmp/test"
			stateFile := filepath.Join(workDir, "deployment-state.yml")
			
			// Setup mock expectations
			if tt.readError != nil {
				ssh.SetError("read "+stateFile, tt.readError)
			} else {
				ssh.SetOutput("read "+stateFile, tt.fileContent)
			}
			
			sm := NewStateManager(ssh, workDir)
			state, err := sm.Load()
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedState.LastKnownGood, state.LastKnownGood)
			}
			
			// Verify read command was attempted
			commands := ssh.GetCommands()
			assert.Contains(t, commands, "read "+stateFile)
		})
	}
}

func TestStateManager_ServiceStateHandling(t *testing.T) {
	t.Run("Service State Updates", func(t *testing.T) {
		ssh := NewMockSSHClient()
		sm := NewStateManager(ssh, "/tmp/test")

		state := DeploymentState{
			LastKnownGood: map[string]ServiceState{
				"service1": {
					Version:      "v1",
					BlueVersion:  "v1",
					GreenVersion: "",
					BlueWeight:   100,
					GreenWeight:  0,
					LastUpdated:  time.Now().Format(time.RFC3339),
				},
			},
		}

		// Set up mock to return the same state when loaded
		stateFile := filepath.Join("/tmp/test", "deployment-state.yml")
		data, err := yaml.Marshal(state)
		assert.NoError(t, err)
		ssh.SetOutput("read "+stateFile, string(data))

		err = sm.Save(state)
		assert.NoError(t, err)

		loadedState, err := sm.Load()
		assert.NoError(t, err)
		assert.Equal(t, state.LastKnownGood["service1"].Version, 
			loadedState.LastKnownGood["service1"].Version)
	})

	t.Run("Service State Validation", func(t *testing.T) {
		ssh := NewMockSSHClient()
		sm := NewStateManager(ssh, "/tmp/test")

		state := DeploymentState{
			LastKnownGood: map[string]ServiceState{
				"service1": {
					BlueWeight:  150,
					GreenWeight: -10,
				},
			},
		}

		err := sm.Save(state)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid weight")
	})
}

func TestStateManager_ConcurrentAccess(t *testing.T) {
	ssh := NewMockSSHClient()
	sm := NewStateManager(ssh, "/tmp/test")

	t.Run("Concurrent Save Operations", func(t *testing.T) {
		var wg sync.WaitGroup
		errChan := make(chan error, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				state := DeploymentState{
					LastKnownGood: map[string]ServiceState{
						fmt.Sprintf("service%d", i): {
							Version:     fmt.Sprintf("v%d", i),
							BlueWeight:  100,
							GreenWeight: 0,
						},
					},
				}
				if err := sm.Save(state); err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Errorf("Concurrent save error: %v", err)
		}
	})
}

func TestStateManager_EdgeCases(t *testing.T) {
	t.Run("Empty State", func(t *testing.T) {
		ssh := NewMockSSHClient()
		sm := NewStateManager(ssh, "/tmp/test")

		state := DeploymentState{}
		err := sm.Save(state)
		assert.NoError(t, err)

		loadedState, err := sm.Load()
		assert.NoError(t, err)
		assert.NotNil(t, loadedState.LastKnownGood)
	})

	t.Run("Invalid Work Directory", func(t *testing.T) {
		ssh := NewMockSSHClient()
		sm := NewStateManager(ssh, "") // Empty work directory

		state := DeploymentState{
			LastKnownGood: map[string]ServiceState{
				"service1": {
					Version:     "v1",
					BlueWeight:  100,
					GreenWeight: 0,
				},
			},
		}

		err := sm.Save(state)
		assert.Error(t, err)
	})

	t.Run("Large State File", func(t *testing.T) {
		ssh := NewMockSSHClient()
		sm := NewStateManager(ssh, "/tmp/test")
		stateFile := filepath.Join("/tmp/test", "deployment-state.yml")

		// Create large state with valid weights
		state := DeploymentState{
			LastKnownGood: make(map[string]ServiceState),
		}
		for i := 0; i < 1000; i++ {
			state.LastKnownGood[fmt.Sprintf("service%d", i)] = ServiceState{
				Version:      fmt.Sprintf("v%d", i),
				BlueVersion:  fmt.Sprintf("v%d", i),
				GreenVersion: fmt.Sprintf("v%d", i+1),
				BlueWeight:   100,
				GreenWeight:  0,
				LastUpdated:  time.Now().Format(time.RFC3339),
			}
		}

		// Set up mock to return the same state when loaded
		data, err := yaml.Marshal(state)
		assert.NoError(t, err)
		ssh.SetOutput("read "+stateFile, string(data))

		err = sm.Save(state)
		assert.NoError(t, err)

		loadedState, err := sm.Load()
		assert.NoError(t, err)
		assert.Equal(t, len(state.LastKnownGood), len(loadedState.LastKnownGood))
	})
}

func TestStateManager_Validation(t *testing.T) {
	tests := []struct {
		name    string
		state   DeploymentState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid state",
			state: DeploymentState{
				LastKnownGood: map[string]ServiceState{
					"service1": {
						Version:     "v1",
						BlueWeight:  60,
						GreenWeight: 40,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid weights sum",
			state: DeploymentState{
				LastKnownGood: map[string]ServiceState{
					"service1": {
						Version:     "v1",
						BlueWeight:  60,
						GreenWeight: 60,
					},
				},
			},
			wantErr: true,
			errMsg:  "weights must sum to 100",
		},
		{
			name: "invalid version format",
			state: DeploymentState{
				LastKnownGood: map[string]ServiceState{
					"service1": {
						BlueVersion:  "v1",
						GreenVersion: "v2",
						Version:      "", // Empty version with blue/green versions set
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := NewStateManager(NewMockSSHClient(), "/tmp/test")
			err := sm.validateState(tt.state)
			
			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
