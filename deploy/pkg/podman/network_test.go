package podman

import (
	"strings"
	"sync"
	"testing"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestNetworkManager(t *testing.T) {
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

	t.Run("Ensure Network", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		mock.SetReturnString("")
		err := manager.EnsureNetwork("web", true)
		assert.NoError(t, err)
	})

	t.Run("Network Already Exists", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		mock.SetReturnString("")
		err := manager.EnsureNetwork("web", false)
		assert.NoError(t, err)
	})

	t.Run("Connect Container to Network", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		mock.SetReturnString("")
		err := manager.ConnectContainer("myapp", []string{"web", "internal"})
		assert.NoError(t, err)
	})

	t.Run("Container Already Connected", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		mock.SetReturnString("myapp")
		err := manager.ConnectContainer("myapp", []string{"web"})
		assert.NoError(t, err)
	})

	t.Run("Disconnect Container", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		mock.SetReturnString("")
		err := manager.DisconnectContainer("myapp", "web")
		assert.NoError(t, err)
	})

	t.Run("Remove Network", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		mock.SetReturnString("")
		err := manager.RemoveNetwork("web")
		assert.NoError(t, err)
	})

	t.Run("List Networks", func(t *testing.T) {
		manager := NewNetworkManager(conn)

		expectedOutput := "bridge\nweb\ninternal"
		mock.SetReturnString(expectedOutput)

		networks, err := manager.ListNetworks()
		assert.NoError(t, err)
		assert.Equal(t, []string{"bridge", "web", "internal"}, networks)
	})
}

func TestNetworkManager_EdgeCases(t *testing.T) {
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

	manager := NewNetworkManager(conn)

	t.Run("Network Name with Special Characters", func(t *testing.T) {
		mock.SetReturnString("not found")
		mock.SetReturnString("")

		networkName := "test-network_123"
		err := manager.EnsureNetwork(networkName, false)
		assert.NoError(t, err)
	})

	t.Run("Empty Network Name", func(t *testing.T) {
		err := manager.EnsureNetwork("", false)
		assert.Error(t, err)
	})
}

func TestNetworkManager_ConcurrentOperations(t *testing.T) {
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

	manager := NewNetworkManager(conn)

	t.Run("Concurrent Network Creation", func(t *testing.T) {
		mock.SetReturnString("not found")
		mock.SetReturnString("")

		var wg sync.WaitGroup
		errChan := make(chan error, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				networkName := "network" + string(rune('0'+i))
				if err := manager.EnsureNetwork(networkName, false); err != nil {
					errChan <- err
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Errorf("Concurrent operation error: %v", err)
		}
	})
}

func TestNetworkManager_Validation(t *testing.T) {
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

	tests := []struct {
		name        string
		networkName string
		internal    bool
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "valid network name",
			networkName: "test-net",
			internal:    false,
			wantErr:     false,
		},
		{
			name:        "invalid characters",
			networkName: "test@network",
			wantErr:     true,
			errMsg:      "invalid network name",
		},
		{
			name:        "name too long",
			networkName: strings.Repeat("a", 64),
			wantErr:     true,
			errMsg:      "network name too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewNetworkManager(conn)
			err := manager.EnsureNetwork(tt.networkName, tt.internal)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
