package pkg

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProxyManager(t *testing.T) {
	t.Run("deploy service", func(t *testing.T) {
		ssh := NewMockSSHClient()
		pm := NewProxyManager(ssh, "kamal-proxy")

		// Set up mock responses for health checks
		ssh.SetOutput("podman inspect --format '{{.State.Status}}' myapp-blue", "running")
		ssh.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-blue", "healthy")

		service := ProxyService{
			Name:        "myapp",
			Version:     "blue",
			Image:       "nginx:1.19",
			Port:        "80",
			Domains:     []string{"app.example.com", "www.example.com"},
			SSL:         true,
			Networks:    []string{"web", "internal"},
			Environment: map[string]string{
				"ENV": "prod",
			},
			Command:     []string{"nginx", "-g", "daemon off;"},
			Weight:      100,
			Labels: map[string]string{
				"custom.label": "value",
			},
			HealthCheckTimeout: 1 * time.Second, // Short timeout for tests
		}

		err := pm.DeployService(service)
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		var deployCmd string
		// Find the podman run command
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman run") {
				deployCmd = cmd
				break
			}
		}

		// Check each required part separately to avoid order dependency
		requiredParts := []string{
			"podman run -d --name myapp-blue",
			"--label 'traefik.enable=true'",
			"--label 'traefik.http.routers.app-example-com.rule=Host(`app.example.com`)'",
			"--label 'traefik.http.routers.app-example-com.tls=true'",
			"--label 'traefik.http.routers.app-example-com.tls.certresolver=default'",
			"--label 'traefik.http.routers.www-example-com.rule=Host(`www.example.com`)'",
			"--label 'traefik.http.routers.www-example-com.tls=true'",
			"--label 'traefik.http.routers.www-example-com.tls.certresolver=default'",
			"--label 'traefik.http.services.myapp.loadbalancer.server.port=80'",
			"--label 'traefik.http.services.myapp.loadbalancer.server.weight=100'",
			"--label 'custom.label=value'",
			"--env 'ENV=prod'",
			"--network web",
			"--network internal",
			"nginx:1.19",
			"nginx -g daemon off;",
		}

		for _, part := range requiredParts {
			assert.Contains(t, deployCmd, part, "Command missing required part: %s", part)
		}
	})

	t.Run("switch traffic", func(t *testing.T) {
		ssh := NewMockSSHClient()
		pm := NewProxyManager(ssh, "kamal-proxy")

		// Set up mock responses for health checks and status checks
		ssh.SetOutput("podman inspect --format '{{.State.Status}}' myapp-blue", "running")
		ssh.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-blue", "healthy")
		ssh.SetOutput("podman inspect --format '{{.State.Status}}' myapp-green", "running")
		ssh.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-green", "healthy")
		ssh.SetOutput("podman update", "") // For weight updates

		err := pm.SwitchTraffic("myapp", "blue", "green", 50, 50)
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		
		// Verify health checks were performed
		foundBlueHealth := false
		foundGreenHealth := false
		foundBlueWeight := false
		foundGreenWeight := false

		for _, cmd := range commands {
			switch {
			case strings.Contains(cmd, "Health.Status") && strings.Contains(cmd, "myapp-blue"):
				foundBlueHealth = true
			case strings.Contains(cmd, "Health.Status") && strings.Contains(cmd, "myapp-green"):
				foundGreenHealth = true
			case strings.Contains(cmd, "weight=50") && strings.Contains(cmd, "myapp-blue"):
				foundBlueWeight = true
			case strings.Contains(cmd, "weight=50") && strings.Contains(cmd, "myapp-green"):
				foundGreenWeight = true
			}
		}

		assert.True(t, foundBlueHealth, "Blue health check not found")
		assert.True(t, foundGreenHealth, "Green health check not found")
		assert.True(t, foundBlueWeight, "Blue weight update not found")
		assert.True(t, foundGreenWeight, "Green weight update not found")
	})

	t.Run("remove version", func(t *testing.T) {
		ssh := NewMockSSHClient()
		pm := NewProxyManager(ssh, "kamal-proxy")

		err := pm.RemoveVersion("myapp", "blue")
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		assert.Contains(t, commands, "podman rm -f myapp-blue")
	})

	t.Run("remove service", func(t *testing.T) {
		ssh := NewMockSSHClient()
		pm := NewProxyManager(ssh, "kamal-proxy")

		err := pm.RemoveService("myapp")
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		assert.Contains(t, commands, "kamal-proxy remove myapp")
	})

	t.Run("handle command failures", func(t *testing.T) {
		ssh := NewMockSSHClient()
		pm := NewProxyManager(ssh, "kamal-proxy")

		// Set up mock responses for health checks
		ssh.SetOutput("podman inspect --format '{{.State.Status}}'", "running")
		ssh.SetOutput("podman inspect --format '{{.State.Health.Status}}'", "healthy")

		// Set up error for deploy
		ssh.SetError("podman run -d --name myapp-blue", fmt.Errorf("deploy failed"))

		service := ProxyService{
			Name:    "myapp",
			Version: "blue",
			Image:   "nginx:1.19",
		}

		err := pm.DeployService(service)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create container")

		// Set up error for traffic switch
		ssh.SetError("podman update", fmt.Errorf("update failed"))
		// Need to set up health check responses for both containers
		ssh.SetOutput("podman inspect --format '{{.State.Status}}' myapp-blue", "running")
		ssh.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-blue", "healthy")
		ssh.SetOutput("podman inspect --format '{{.State.Status}}' myapp-green", "running")
		ssh.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-green", "healthy")

		err = pm.SwitchTraffic("myapp", "blue", "green", 50, 50)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")

		// Set up error for remove
		ssh.SetError("podman rm -f", fmt.Errorf("remove failed"))
		err = pm.RemoveVersion("myapp", "blue")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "remove failed")
	})
}

func TestProxyManagerHelpers(t *testing.T) {
	pm := NewProxyManager(nil, "")

	t.Run("build labels argument", func(t *testing.T) {
		labels := map[string]string{
			"label1": "value1",
			"label2": "value2",
		}
		result := pm.buildLabelsArg(labels)
		assert.Contains(t, result, "--label 'label1=value1'")
		assert.Contains(t, result, "--label 'label2=value2'")
	})

	t.Run("build env argument", func(t *testing.T) {
		env := map[string]string{
			"ENV1": "value1",
			"ENV2": "value2",
		}
		result := pm.buildEnvArg(env)
		assert.Contains(t, result, "--env 'ENV1=value1'")
		assert.Contains(t, result, "--env 'ENV2=value2'")
	})

	t.Run("build networks argument", func(t *testing.T) {
		networks := []string{"net1", "net2"}
		result := pm.buildNetworksArg(networks)
		assert.Equal(t, "--network net1 --network net2", result)
	})

	t.Run("empty arguments", func(t *testing.T) {
		assert.Equal(t, "", pm.buildLabelsArg(nil))
		assert.Equal(t, "", pm.buildEnvArg(nil))
		assert.Equal(t, "", pm.buildNetworksArg(nil))
	})
}

func TestProxyManager_HealthCheck(t *testing.T) {
	mock := NewMockSSHClient()
	pm := NewProxyManager(mock, "kamal-proxy")

	t.Run("Wait for Healthy Container", func(t *testing.T) {
		mock.SetOutput("podman run", "")
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-v1", "healthy")

		service := ProxyService{
			Name:              "myapp",
			Version:          "v1",
			Image:            "nginx:latest",
			HealthCheckTimeout: 5 * time.Second,
		}

		err := pm.DeployService(service)
		assert.NoError(t, err)

		commands := mock.GetCommands()
		foundHealthCheck := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman inspect --format '{{.State.Health.Status}}' myapp-v1") {
				foundHealthCheck = true
				break
			}
		}
		assert.True(t, foundHealthCheck, "Health check command not found")
	})

	t.Run("Container Never Becomes Healthy", func(t *testing.T) {
		mock := NewMockSSHClient()
		pm := NewProxyManager(mock, "kamal-proxy")

		// Set up mock responses
		mock.SetOutput("podman run -d --name myapp-v1", "")
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-v1", "unhealthy")
		mock.SetOutput("podman rm -f myapp-v1", "")

		service := ProxyService{
			Name:              "myapp",
			Version:          "v1",
			Image:            "nginx:latest",
			HealthCheckTimeout: 300 * time.Millisecond, // Very short timeout for tests
		}

		startTime := time.Now()
		err := pm.DeployService(service)
		duration := time.Since(startTime)

		// Verify error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container failed health check")

		// Verify timing
		assert.True(t, duration >= 300*time.Millisecond, "Test completed too quickly")
		assert.True(t, duration < 600*time.Millisecond, "Test took too long")

		// Verify cleanup was attempted
		commands := mock.GetCommands()
		foundCleanup := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman rm -f myapp-v1") {
				foundCleanup = true
				break
			}
		}
		assert.True(t, foundCleanup, "Cleanup command not found")
	})
}

func TestProxyManager_TrafficSwitching(t *testing.T) {
	mock := NewMockSSHClient()
	pm := NewProxyManager(mock, "kamal-proxy")

	t.Run("Gradual Traffic Switch", func(t *testing.T) {
		// Setup initial state with specific container names
		mock.SetOutput("podman inspect --format '{{.State.Status}}' myapp-v1", "running")
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-v1", "healthy")
		mock.SetOutput("podman inspect --format '{{.State.Status}}' myapp-v2", "running")
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}' myapp-v2", "healthy")
		mock.SetOutput("podman update --label-add 'traefik.http.services.myapp.loadbalancer.server.weight=75' myapp-v1", "")
		mock.SetOutput("podman update --label-add 'traefik.http.services.myapp.loadbalancer.server.weight=25' myapp-v2", "")

		err := pm.SwitchTraffic("myapp", "v1", "v2", 75, 25)
		assert.NoError(t, err)

		commands := mock.GetCommands()
		assert.Contains(t, commands, 
			"podman update --label-add 'traefik.http.services.myapp.loadbalancer.server.weight=75' myapp-v1")
		assert.Contains(t, commands, 
			"podman update --label-add 'traefik.http.services.myapp.loadbalancer.server.weight=25' myapp-v2")
	})

	t.Run("Rollback on Unhealthy New Version", func(t *testing.T) {
		mock := NewMockSSHClient()
		pm := NewProxyManager(mock, "kamal-proxy")
		pm.healthCheckTimeout = 100 * time.Millisecond
		pm.healthCheckInterval = 50 * time.Millisecond

		// Set up mock responses for status checks
		mock.SetOutput("podman inspect --format '{{.State.Status}}' myapp-v1", "running")
		mock.SetOutput("podman inspect --format '{{.State.Status}}' myapp-v2", "running")

		// Set up mock responses for health checks with a custom handler
		mock.SetCustomHandler("podman inspect", func(cmd string) (string, error) {
			if strings.Contains(cmd, "Health.Status") {
				if strings.Contains(cmd, "myapp-v1") {
					return "healthy", nil
				}
				if strings.Contains(cmd, "myapp-v2") {
					return "unhealthy", nil
				}
			}
			if strings.Contains(cmd, "State.Status") {
				return "running", nil
			}
			return "", fmt.Errorf("unexpected command: %s", cmd)
		})

		// Set up mock responses for weight updates
		mock.SetOutput("podman update --label-add 'traefik.http.services.myapp.loadbalancer.server.weight=50' myapp-v2", "")
		mock.SetOutput("podman update --label-add 'traefik.http.services.myapp.loadbalancer.server.weight=0' myapp-v2", "")

		// Run with shorter timeout for test
		testTimeout := 500 * time.Millisecond
		ctx, cancel := context.WithTimeout(context.Background(), testTimeout)
		defer cancel()

		errChan := make(chan error, 1)
		go func() {
			errChan <- pm.SwitchTraffic("myapp", "v1", "v2", 50, 50)
		}()

		select {
		case err := <-errChan:
			if err == nil {
				t.Error("Expected error for unhealthy new version")
			}
			if !strings.Contains(err.Error(), "green version not healthy") {
				t.Errorf("Expected 'green version not healthy' error, got: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	})
}

func TestProxyManager_ComplexScenarios(t *testing.T) {
	mock := NewMockSSHClient()
	pm := NewProxyManager(mock, "kamal-proxy")

	t.Run("Multiple Domain Configuration", func(t *testing.T) {
		service := ProxyService{
			Name:    "myapp",
			Version: "v1",
			Image:   "nginx:latest",
			Domains: []string{
				"example.com",
				"www.example.com",
				"api.example.com",
			},
			SSL: true,
		}

		err := pm.DeployService(service)
		assert.NoError(t, err)

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]

		// Verify all domains are configured
		for _, domain := range service.Domains {
			assert.Contains(t, lastCmd, fmt.Sprintf("Host(`%s`)", domain))
			assert.Contains(t, lastCmd, "tls=true")
			assert.Contains(t, lastCmd, "certresolver=default")
		}
	})

	t.Run("Network Isolation", func(t *testing.T) {
		service := ProxyService{
			Name:     "myapp",
			Version:  "v1",
			Image:    "nginx:latest",
			Networks: []string{"frontend", "backend"},
			Private:  true,
		}

		err := pm.DeployService(service)
		assert.NoError(t, err)

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]

		// Verify network configuration
		assert.Contains(t, lastCmd, "--network frontend")
		assert.Contains(t, lastCmd, "--network backend")
	})
}

func TestProxyManager_ErrorHandling(t *testing.T) {
	mock := NewMockSSHClient()
	pm := NewProxyManager(mock, "kamal-proxy")

	t.Run("Invalid Domain Format", func(t *testing.T) {
		service := ProxyService{
			Name:    "myapp",
			Version: "v1",
			Domains: []string{"invalid domain"},
		}

		err := pm.DeployService(service)
		assert.Error(t, err)
	})

	t.Run("Container Creation Failure", func(t *testing.T) {
		mock.SetError("podman run", fmt.Errorf("network not found"))

		service := ProxyService{
			Name:    "myapp",
			Version: "v1",
			Image:   "nginx:latest",
		}

		err := pm.DeployService(service)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create container")
	})
}

func TestProxyManager_Validation(t *testing.T) {
	tests := []struct {
		name    string
		service ProxyService
		wantErr bool
		errMsg  string
	}{
		{
			name: "missing name",
			service: ProxyService{
				Version: "v1",
				Image:   "nginx:latest",
			},
			wantErr: true,
			errMsg:  "service name is required",
		},
		{
			name: "missing version",
			service: ProxyService{
				Name:  "myapp",
				Image: "nginx:latest",
			},
			wantErr: true,
			errMsg:  "version is required",
		},
		{
			name: "invalid port",
			service: ProxyService{
				Name:    "myapp",
				Version: "v1",
				Port:    "invalid",
			},
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name: "invalid weight",
			service: ProxyService{
				Name:    "myapp",
				Version: "v1",
				Weight:  101,
			},
			wantErr: true,
			errMsg:  "weight must be between 0 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := NewProxyManager(NewMockSSHClient(), "kamal-proxy")
			err := pm.DeployService(tt.service)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
