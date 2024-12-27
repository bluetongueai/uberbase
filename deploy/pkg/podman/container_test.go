package podman

import (
	"testing"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestContainerManager(t *testing.T) {
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

	t.Run("Run Basic Container", func(t *testing.T) {
		manager := NewContainerManager(conn)

		container := Container{
			Name:  "web",
			Image: "nginx:latest",
		}

		mock.SetReturnString("running")

		err := manager.Create(container)
		assert.NoError(t, err)
	})

	t.Run("Run Container with All Options", func(t *testing.T) {
		manager := NewContainerManagerWithVolumes(conn, NewVolumeManager(conn))

		container := Container{
			Name:    "web",
			Image:   "nginx:latest",
			Command: []string{"nginx", "-g", "daemon off;"},
			User:    "nginx",
			Ports:   []string{"80:80", "443:443"},
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
			Healthcheck: &Healthcheck{
				Test:     []string{"CMD-SHELL", "curl -f http://localhost/ || exit 1"},
				Interval: "30s",
				Timeout:  "10s",
				Retries:  3,
			},
		}

		mock.SetReturnString("")
		mock.SetReturnString("healthy")

		err := manager.Create(container)
		assert.NoError(t, err)
	})

	t.Run("Container with Healthcheck", func(t *testing.T) {
		manager := NewContainerManager(conn)

		container := Container{
			Name:  "web",
			Image: "nginx",
			Healthcheck: &Healthcheck{
				Test:     []string{"CMD", "curl", "-f", "http://localhost/"},
				Interval: "5s",
				Timeout:  "3s",
				Retries:  2,
			},
		}

		mock.SetReturnString("")
		mock.SetReturnString("healthy")

		err := manager.Create(container)
		assert.NoError(t, err)
	})

	t.Run("Remove Container", func(t *testing.T) {
		manager := NewContainerManager(conn)

		mock.SetReturnString("")

		err := manager.Remove("web")
		assert.NoError(t, err)
	})

	t.Run("Run Container Error", func(t *testing.T) {
		manager := NewContainerManager(conn)

		container := Container{
			Name:  "web",
			Image: "nginx",
		}

		mock.SetReturnString("container already exists")

		err := manager.Create(container)
		assert.Error(t, err)
	})
}

func TestContainerManager_DependencyHandling(t *testing.T) {
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

	manager := NewContainerManager(conn)

	t.Run("Wait for Healthy Dependency", func(t *testing.T) {
		mock.SetReturnString("healthy")
		mock.SetReturnString("")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			DependsOn: map[string]DependsOnConfig{
				"db": {
					Condition: "service_healthy",
				},
			},
		}

		err := manager.Create(container)
		assert.NoError(t, err)
	})

	t.Run("Wait for Started Dependency", func(t *testing.T) {
		mock.SetReturnString("running")
		mock.SetReturnString("")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			DependsOn: map[string]DependsOnConfig{
				"cache": {
					Condition: "service_started",
				},
			},
		}

		err := manager.Create(container)
		assert.NoError(t, err)
	})

	t.Run("Dependency Timeout", func(t *testing.T) {
		manager := NewContainerManager(conn).WithTimeouts(2, 100)

		mock.SetReturnString("unhealthy")

		container := Container{
			Name:  "app",
			Image: "app:latest",
			DependsOn: map[string]DependsOnConfig{
				"db": {
					Condition: "service_healthy",
				},
			},
		}

		err := manager.Create(container)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "container failed to become healthy within")
	})
}

func TestContainerManager_ResourceConfigs(t *testing.T) {
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

	manager := NewContainerManager(conn)

	t.Run("Container with Resource Limits", func(t *testing.T) {
		mock.SetReturnString("")

		container := Container{
			Name:  "app",
			Image: "app:latest",
		}

		err := manager.Create(container)
		assert.NoError(t, err)
	})
}

func TestContainerManager_NetworkConfigs(t *testing.T) {
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

	manager := NewContainerManager(conn)

	t.Run("Container with Multiple Networks", func(t *testing.T) {
		mock.SetReturnString("")

		container := Container{
			Name:      "app",
			Image:     "app:latest",
			Networks:  []string{"frontend", "backend"},
			DNS:       []string{"8.8.8.8"},
			DNSSearch: []string{"example.com"},
		}

		err := manager.Create(container)
		assert.NoError(t, err)
	})
}

func TestContainerManager_SecurityConfigs(t *testing.T) {
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

	manager := NewContainerManager(conn)

	t.Run("Container with Security Options", func(t *testing.T) {
		mock.SetReturnString("")

		container := Container{
			Name:        "app",
			Image:       "app:latest",
			SecurityOpt: []string{"no-new-privileges"},
			Privileged:  true,
			ReadOnly:    true,
		}

		err := manager.Create(container)
		assert.NoError(t, err)
	})
}

func TestContainerManager_Cleanup(t *testing.T) {
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

	manager := NewContainerManager(conn)

	t.Run("Cleanup on Failed Run", func(t *testing.T) {
		mock.SetReturnString("")
		mock.SetReturnString("run failed")
		mock.SetReturnString("")

		container := Container{
			Name:    "app",
			Image:   "app:latest",
			Volumes: []string{"data:/data"},
		}

		err := manager.Create(container)
		assert.Error(t, err)
	})
}
