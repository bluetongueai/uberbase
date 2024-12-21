package pkg

import (
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestNetworkManager(t *testing.T) {
	t.Run("Ensure Network", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		// Network doesn't exist
		mock.SetError("podman network ls | grep -q web", fmt.Errorf("not found"))
		mock.SetOutput("podman network create web --internal --driver bridge", "")

		if err := manager.EnsureNetwork("web", true); err != nil {
			t.Errorf("EnsureNetwork failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmds := []string{
			"podman network ls | grep -q web",
			"podman network create web --internal --driver bridge",
		}
		for i, cmd := range commands {
			if cmd != expectedCmds[i] {
				t.Errorf("Expected command %q, got %q", expectedCmds[i], cmd)
			}
		}
	})

	t.Run("Network Already Exists", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		mock.SetOutput("podman network ls | grep -q web", "")

		if err := manager.EnsureNetwork("web", false); err != nil {
			t.Errorf("EnsureNetwork failed: %v", err)
		}

		commands := mock.GetCommands()
		if len(commands) != 1 || commands[0] != "podman network ls | grep -q web" {
			t.Errorf("Unexpected commands: %v", commands)
		}
	})

	t.Run("Connect Container to Network", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		// Container not connected
		mock.SetError("podman network inspect web | grep -q myapp", fmt.Errorf("not connected"))
		mock.SetError("podman network inspect internal | grep -q myapp", fmt.Errorf("not connected"))
		mock.SetOutput("podman network connect web myapp", "")
		mock.SetOutput("podman network connect internal myapp", "")

		if err := manager.ConnectContainer("myapp", []string{"web", "internal"}); err != nil {
			t.Errorf("ConnectContainer failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmds := []string{
			"podman network inspect web | grep -q myapp",
			"podman network connect web myapp",
			"podman network inspect internal | grep -q myapp",
			"podman network connect internal myapp",
		}
		for i, cmd := range commands {
			if cmd != expectedCmds[i] {
				t.Errorf("Expected command %q, got %q", expectedCmds[i], cmd)
			}
		}
	})

	t.Run("Container Already Connected", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		mock.SetOutput("podman network inspect web | grep -q myapp", "myapp")

		if err := manager.ConnectContainer("myapp", []string{"web"}); err != nil {
			t.Errorf("ConnectContainer failed: %v", err)
		}

		commands := mock.GetCommands()
		if len(commands) != 1 || commands[0] != "podman network inspect web | grep -q myapp" {
			t.Errorf("Unexpected commands: %v", commands)
		}
	})

	t.Run("Disconnect Container", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		mock.SetOutput("podman network disconnect", "")

		if err := manager.DisconnectContainer("myapp", "web"); err != nil {
			t.Errorf("DisconnectContainer failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman network disconnect web myapp"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("Remove Network", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		mock.SetOutput("podman network rm", "")

		if err := manager.RemoveNetwork("web"); err != nil {
			t.Errorf("RemoveNetwork failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman network rm web"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("List Networks", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		expectedOutput := "bridge\nweb\ninternal"
		mock.SetOutput("podman network ls", expectedOutput)

		networks, err := manager.ListNetworks()
		if err != nil {
			t.Errorf("ListNetworks failed: %v", err)
		}

		if len(networks) != 3 {
			t.Errorf("Expected 3 networks, got %d", len(networks))
		}

		expectedNetworks := []string{"bridge", "web", "internal"}
		for i, net := range networks {
			if net != expectedNetworks[i] {
				t.Errorf("Expected network %q, got %q", expectedNetworks[i], net)
			}
		}
	})
}

func TestNetworkManager_EdgeCases(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewNetworkManager(mock)

	t.Run("Network Name with Special Characters", func(t *testing.T) {
		mock.SetError("podman network ls", fmt.Errorf("not found"))
		mock.SetOutput("podman network create", "")

		networkName := "test-network_123"
		if err := manager.EnsureNetwork(networkName, false); err != nil {
			t.Errorf("Failed to handle special characters: %v", err)
		}

		commands := mock.GetCommands()
		for _, cmd := range commands {
			if strings.Contains(cmd, "create") && !strings.Contains(cmd, networkName) {
				t.Errorf("Network name not properly handled in command: %s", cmd)
			}
		}
	})

	t.Run("Empty Network Name", func(t *testing.T) {
		err := manager.EnsureNetwork("", false)
		if err == nil {
			t.Error("Expected error for empty network name")
		}
	})
}

func TestNetworkManager_ConcurrentOperations(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewNetworkManager(mock)

	t.Run("Concurrent Network Creation", func(t *testing.T) {
		mock.SetError("podman network ls", fmt.Errorf("not found"))
		mock.SetOutput("podman network create", "")

		var wg sync.WaitGroup
		errChan := make(chan error, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				networkName := fmt.Sprintf("network%d", i)
				if err := manager.EnsureNetwork(networkName, false); err != nil {
					errChan <- fmt.Errorf("network %s: %v", networkName, err)
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

func TestNetworkManager_ErrorHandling(t *testing.T) {
	mock := NewMockSSHClient()
	manager := NewNetworkManager(mock)

	t.Run("Network Create Permission Denied", func(t *testing.T) {
		mock.SetError("podman network ls | grep -q test", fmt.Errorf("not found"))
		mock.SetError("podman network create test --driver bridge", fmt.Errorf("permission denied"))

		err := manager.EnsureNetwork("test", false)
		if err == nil {
			t.Error("Expected permission error")
		}
		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("Expected permission denied error, got: %v", err)
		}
	})

	t.Run("Network Already Exists but Inspect Fails", func(t *testing.T) {
		mock.SetOutput("podman network ls | grep -q test", "")  // Network exists
		mock.SetError("podman network inspect test | grep -q container1", fmt.Errorf("inspect failed"))
		mock.SetError("podman network connect test container1", fmt.Errorf("connect failed"))

		err := manager.ConnectContainer("container1", []string{"test"})
		if err == nil {
			t.Error("Expected error when inspect fails")
		}
		if !strings.Contains(err.Error(), "connect failed") {
			t.Errorf("Expected connect failed error, got: %v", err)
		}
	})
}

func TestNetworkManager_Validation(t *testing.T) {
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
			manager := NewNetworkManager(NewMockSSHClient())
			err := manager.EnsureNetwork(tt.networkName, tt.internal)
			
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

func TestNetworkManager_ComplexOperations(t *testing.T) {
	t.Run("Multiple Network Operations", func(t *testing.T) {
		mock := NewMockSSHClient()
		manager := NewNetworkManager(mock)

		// Setup mock responses in correct order
		mock.SetError("podman network ls | grep -q test-net", fmt.Errorf("not found"))
		mock.SetOutput("podman network create test-net --internal --driver bridge", "")
		mock.SetError("podman network inspect test-net | grep -q container1", fmt.Errorf("not found"))
		mock.SetOutput("podman network connect test-net container1", "")
		mock.SetOutput("podman network disconnect test-net container1", "")
		mock.SetOutput("podman network rm test-net", "")

		// Create network
		if err := manager.EnsureNetwork("test-net", true); err != nil {
			t.Fatalf("Failed to create network: %v", err)
		}

		// Connect container
		if err := manager.ConnectContainer("container1", []string{"test-net"}); err != nil {
			t.Fatalf("Failed to connect container: %v", err)
		}

		// Disconnect container
		if err := manager.DisconnectContainer("container1", "test-net"); err != nil {
			t.Fatalf("Failed to disconnect container: %v", err)
		}

		// Remove network
		if err := manager.RemoveNetwork("test-net"); err != nil {
			t.Fatalf("Failed to remove network: %v", err)
		}

		// Verify command sequence
		commands := mock.GetCommands()
		expectedSequence := []string{
			"podman network ls | grep -q test-net",
			"podman network create test-net --internal --driver bridge",
			"podman network inspect test-net | grep -q container1",
			"podman network connect test-net container1",
			"podman network disconnect test-net container1",
			"podman network rm test-net",
		}

		// Verify each command matches expected sequence
		for i, expected := range expectedSequence {
			if i >= len(commands) {
				t.Errorf("Missing command at position %d, expected %q", i, expected)
				continue
			}
			if commands[i] != expected {
				t.Errorf("Command mismatch at position %d\nExpected: %q\nGot: %q", 
					i, expected, commands[i])
			}
		}
	})
}
