package pkg

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestContainerManager(t *testing.T) {
	t.Run("Run Basic Container", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock)

		container := Container{
			Name:  "web",
			Image: "nginx:latest",
		}

		mock.SetOutput("podman run", "")

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman run -d --name web nginx:latest"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("Run Container with All Options", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock)

		container := Container{
			Name:  "web",
			Image: "nginx:latest",
			Command: []string{"nginx", "-g", "daemon off;"},
			User:   "nginx",
			Ports:  []string{"80:80", "443:443"},
			Volumes: []string{
				"nginx_data:/var/www",
				"/etc/nginx:/etc/nginx:ro",
			},
			Environment: map[string]string{
				"NGINX_HOST": "example.com",
				"NGINX_PORT": "80",
			},
			EnvFile:      []string{"/etc/nginx/env"},
			Capabilities: []string{"NET_ADMIN", "SYS_TIME"},
			ExtraHosts:   []string{"host.docker.internal:host-gateway"},
			Restart:      "unless-stopped",
			Healthcheck: Healthcheck{
				Test:     []string{"CMD-SHELL", "curl -f http://localhost/ || exit 1"},
				Interval: "30s",
				Timeout:  "10s",
				Retries:  3,
			},
		}

		mock.SetOutput("podman volume create", "")
		mock.SetOutput("podman run", "")
		mock.SetOutput("podman inspect", "healthy") // For healthcheck

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman run -d --name web --user nginx --restart unless-stopped " +
			"--cap-add NET_ADMIN --cap-add SYS_TIME " +
			"--add-host host.docker.internal:host-gateway " +
			"-p 80:80 -p 443:443 " +
			"-v nginx_data:/var/www -v /etc/nginx:/etc/nginx:ro " +
			"--env-file /etc/nginx/env " +
			"-e NGINX_HOST=example.com -e NGINX_PORT=80 " +
			"--health-cmd \"CMD-SHELL curl -f http://localhost/ || exit 1\" " +
			"--health-interval 30s --health-timeout 10s --health-retries 3 " +
			"nginx:latest nginx -g daemon off;"

		var runCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "podman run") {
				runCmd = cmd
				break
			}
		}

		if runCmd == "" {
			t.Fatal("Run command not found in executed commands")
		}

		if runCmd != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, runCmd)
		}
	})

	t.Run("Container with Volume Creation", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock)

		container := Container{
			Name:    "db",
			Image:   "postgres",
			Volumes: []string{"pgdata:/var/lib/postgresql/data"},
		}

		mock.SetCustomHandler("podman", func(cmd string) (string, error) {
			if strings.Contains(cmd, "volume inspect") {
				return "", fmt.Errorf("not found")
			}
			if strings.Contains(cmd, "volume create") {
				return "", nil
			}
			if strings.Contains(cmd, "run") {
				return "", nil
			}
			return "", fmt.Errorf("unexpected command: %s", cmd)
		})

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		foundVolumeCmd := false
		foundRunCmd := false
		
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman volume") {
				foundVolumeCmd = true
			}
			if strings.Contains(cmd, "podman run") {
				foundRunCmd = true
			}
		}

		if !foundVolumeCmd {
			t.Error("Volume command not found")
		}
		if !foundRunCmd {
			t.Error("Run command not found")
		}
	})

	t.Run("Container with Healthcheck", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock)

		container := Container{
			Name:  "web",
			Image: "nginx",
			Healthcheck: Healthcheck{
				Test:     []string{"CMD", "curl", "-f", "http://localhost/"},
				Interval: "5s",
				Timeout:  "3s",
				Retries:  2,
			},
		}

		mock.SetOutput("podman run", "")
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}'", "healthy")

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		if len(commands) < 1 {
			t.Fatal("Expected at least 1 command")
		}

		runCmd := commands[0]
		expectedFlags := []string{
			"--health-cmd",
			"--health-interval 5s",
			"--health-timeout 3s",
			"--health-retries 2",
		}
		for _, flag := range expectedFlags {
			if !strings.Contains(runCmd, flag) {
				t.Errorf("Expected %q in run command, got: %s", flag, runCmd)
			}
		}
	})

	t.Run("Remove Container", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock)

		mock.SetOutput("podman rm", "")

		if err := manager.Remove("web"); err != nil {
			t.Errorf("Remove failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman rm -f web"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("Run Container Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock)

		container := Container{
			Name:  "web",
			Image: "nginx",
		}

		mock.SetError("podman run", fmt.Errorf("container already exists"))

		if err := manager.Run(container); err == nil {
			t.Error("Expected error when container already exists")
		}
	})
}

func TestContainerManager_DependencyHandling(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewContainerManager(mock)

	t.Run("Wait for Healthy Dependency", func(t *testing.T) {
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}'", "healthy")
		mock.SetOutput("podman run", "")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			DependsOn: []DependencyConfig{
				{
					Service:   "db",
					Condition: "service_healthy",
				},
			},
		}

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		foundHealthCheck := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman inspect --format '{{.State.Health.Status}}' db") {
				foundHealthCheck = true
				break
			}
		}
		if !foundHealthCheck {
			t.Error("Health check command not found")
		}
	})

	t.Run("Wait for Started Dependency", func(t *testing.T) {
		mock.SetOutput("podman inspect --format '{{.State.Status}}'", "running")
		mock.SetOutput("podman run", "")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			DependsOn: []DependencyConfig{
				{
					Service:   "cache",
					Condition: "service_started",
				},
			},
		}

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}
	})

	t.Run("Dependency Timeout", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewContainerManager(mock).WithTimeouts(2, 100*time.Millisecond)
		
		// Set up mock to always return unhealthy
		mock.SetOutput("podman inspect --format '{{.State.Health.Status}}'", "unhealthy")
		
		container := Container{
			Name:  "app",
			Image: "app:latest",
			DependsOn: []DependencyConfig{
				{
					Service:   "db",
					Condition: "service_healthy",
				},
			},
		}

		err := manager.Run(container)
		if err == nil {
			t.Error("Expected timeout error waiting for dependency")
		}
		
		// Verify the error message
		expectedErr := "container failed to become healthy within"
		if !strings.Contains(err.Error(), expectedErr) {
			t.Errorf("Expected error containing %q, got %v", expectedErr, err)
		}
		
		// Verify number of health checks
		commands := mock.GetCommands()
		healthChecks := 0
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman inspect --format '{{.State.Health.Status}}' db") {
				healthChecks++
			}
		}
		
		// Should have exactly 2 health checks (maxRetries)
		if healthChecks != 2 {
			t.Errorf("Expected 2 health checks, got %d", healthChecks)
		}
	})
}

func TestContainerManager_ResourceConfigs(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewContainerManager(mock)

	t.Run("Container with Resource Limits", func(t *testing.T) {
		mock.SetOutput("podman run", "")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			Deploy: DeployConfig{
				Resources: ResourceConfig{
					Limits: Resources{
						CPUs:   "0.5",
						Memory: "512m",
					},
					Reservations: Resources{
						CPUs:   "0.1",
						Memory: "128m",
					},
				},
			},
		}

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		expectedFlags := []string{
			"--cpus 0.5",
			"--memory 512m",
		}
		for _, flag := range expectedFlags {
			if !strings.Contains(lastCmd, flag) {
				t.Errorf("Expected command to contain %q, got: %s", flag, lastCmd)
			}
		}
	})
}

func TestContainerManager_NetworkConfigs(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewContainerManager(mock)

	t.Run("Container with Multiple Networks", func(t *testing.T) {
		mock.SetOutput("podman run", "")
		mock.SetOutput("podman network connect", "")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			Networks: []string{"frontend", "backend"},
			DNS: []string{"8.8.8.8"},
			DNSSearch: []string{"example.com"},
		}

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		foundNetworkConnect := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman network connect backend app") {
				foundNetworkConnect = true
				break
			}
		}
		if !foundNetworkConnect {
			t.Error("Network connect command not found")
		}
	})
}

func TestContainerManager_SecurityConfigs(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewContainerManager(mock)

	t.Run("Container with Security Options", func(t *testing.T) {
		mock.SetOutput("podman run", "")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			SecurityOpt: []string{"no-new-privileges"},
			Privileged: true,
			ReadOnly:   true,
		}

		if err := manager.Run(container); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		lastCmd := commands[len(commands)-1]
		expectedFlags := []string{
			"--security-opt no-new-privileges",
			"--privileged",
			"--read-only",
		}
		for _, flag := range expectedFlags {
			if !strings.Contains(lastCmd, flag) {
				t.Errorf("Expected command to contain %q, got: %s", flag, lastCmd)
			}
		}
	})
}

func TestContainerManager_Cleanup(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewContainerManager(mock)

	t.Run("Cleanup on Failed Run", func(t *testing.T) {
		mock.SetOutput("podman volume create", "")
		mock.SetError("podman run", fmt.Errorf("run failed"))
		mock.SetOutput("podman rm", "")

		container := Container{
			Name:    "app",
			Image:   "app:latest",
			Volumes: []string{"data:/data"},
		}

		err := manager.Run(container)
		if err == nil {
			t.Error("Expected run to fail")
		}

		// Verify cleanup was attempted
		commands := mock.GetCommands()
		foundCleanup := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman rm -f app") {
				foundCleanup = true
				break
			}
		}
		if !foundCleanup {
			t.Error("Cleanup command not found after failed run")
		}
	})
}

func TestContainerManager_Validation(t *testing.T) {
	tests := []struct {
		name      string
		container Container
		wantErr   bool
		errMsg    string
	}{
		{
			name: "invalid memory format",
			container: Container{
				Name:  "app",
				Image: "app:latest",
				Deploy: DeployConfig{
					Resources: ResourceConfig{
						Limits: Resources{
							Memory: "invalid",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid memory format",
		},
		{
			name: "invalid CPU format",
			container: Container{
				Name:  "app",
				Image: "app:latest",
				Deploy: DeployConfig{
					Resources: ResourceConfig{
						Limits: Resources{
							CPUs: "invalid",
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "invalid CPU format",
		},
		// Add more validation test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewContainerManager(NewMockSSHClient())
			err := manager.Run(tt.container)
			
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
