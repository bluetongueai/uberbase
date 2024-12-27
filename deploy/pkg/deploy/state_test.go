package deploy

import (
	"testing"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNewStateManager(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	workDir := "/tmp/test"
	sm := NewStateManager(conn, workDir)

	assert.NotNil(t, sm)
	assert.Equal(t, conn, sm.ssh)
	assert.Equal(t, workDir, sm.workDir)
}

func TestSave(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	workDir := "/tmp/test"
	sm := NewStateManager(conn, workDir)

	tests := []struct {
		name          string
		state         DeploymentState
		expectedError bool
	}{
		{
			name: "successful save",
			state: DeploymentState{
				Services: map[string]*ServiceState{
					"service1": {
						BlueVersion: "nginx:latest",
						BlueWeight:  100,
						GreenWeight: 0,
					},
				},
			},
			expectedError: false,
		},
		{
			name: "empty state",
			state: DeploymentState{
				Services: make(map[string]*ServiceState),
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.Save(tt.state)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	workDir := "/tmp/test"
	sm := NewStateManager(conn, workDir)

	tests := []struct {
		name          string
		fileContent   string
		expectedState DeploymentState
		expectedError bool
	}{
		{
			name: "successful load",
			fileContent: `last_known_good:
  service1:
    version: "nginx:latest"
    blue_weight: 100
    green_weight: 0
    blue_version: ""
    green_version: ""
    last_updated: ""`,
			expectedState: DeploymentState{
				Services: map[string]*ServiceState{
					"service1": {
						BlueVersion: "nginx:latest",
						BlueWeight:  100,
						GreenWeight: 0,
					},
				},
			},
			expectedError: false,
		},
		{
			name:        "invalid yaml",
			fileContent: "invalid: yaml: content",
			expectedState: DeploymentState{
				Services: nil,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.SetReturnString(tt.fileContent)

			state, err := sm.Load()

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedState.Services, state.Services)
			}
		})
	}
}

func TestStateManager_ServiceStateHandling(t *testing.T) {
	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	sm := NewStateManager(conn, "/tmp/test")

	t.Run("Service State Updates", func(t *testing.T) {
		state := DeploymentState{
			Services: map[string]*ServiceState{
				"service1": {
					BlueVersion:  "v1",
					GreenVersion: "",
					BlueWeight:   100,
					GreenWeight:  0,
					LastUpdated:  time.Now().Format(time.RFC3339),
				},
			},
		}

		err := sm.Save(state)
		assert.NoError(t, err)

		loadedState, err := sm.Load()
		assert.NoError(t, err)
		assert.Equal(t, state.Services["service1"].BlueVersion, loadedState.Services["service1"].BlueVersion)
	})

	t.Run("Service State Validation", func(t *testing.T) {
		state := DeploymentState{
			Services: map[string]*ServiceState{
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

func TestStateManager_Validation(t *testing.T) {
	conn, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	sm := NewStateManager(conn, "/tmp/test")

	tests := []struct {
		name    string
		state   DeploymentState
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid state",
			state: DeploymentState{
				Services: map[string]*ServiceState{
					"service1": {
						BlueVersion:  "v1",
						GreenVersion: "v2",
						BlueWeight:   60,
						GreenWeight:  40,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid weights sum",
			state: DeploymentState{
				Services: map[string]*ServiceState{
					"service1": {
						BlueVersion:  "v1",
						GreenVersion: "v2",
						BlueWeight:   60,
						GreenWeight:  60,
					},
				},
			},
			wantErr: true,
			errMsg:  "weights must sum to 100",
		},
		{
			name: "invalid version format",
			state: DeploymentState{
				Services: map[string]*ServiceState{
					"service1": {
						BlueVersion:  "v1",
						GreenVersion: "v2",
						BlueWeight:   100,
						GreenWeight:  0,
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid version",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.validateState(tt.state)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
