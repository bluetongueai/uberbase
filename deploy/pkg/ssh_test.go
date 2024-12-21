package pkg

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testKey = `-----BEGIN OPENSSH PRIVATE KEY-----
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
-----END OPENSSH PRIVATE KEY-----`

type MockSession struct {
	output []byte
	err    error
}

func (m *MockSession) CombinedOutput(cmd string) ([]byte, error) {
	return m.output, m.err
}

func (m *MockSession) Close() error {
	return nil
}

func TestSSHClient(t *testing.T) {
	t.Run("New Client", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", testKey)
		if client.Host != "host" {
			t.Error("Host not set correctly")
		}
		if client.User != "user" {
			t.Error("User not set correctly")
		}
		if client.Port != 22 {
			t.Error("Port not set correctly")
		}
		if client.keyData != testKey {
			t.Error("Key data not set correctly")
		}
	})

	t.Run("Connect With Invalid Key", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", "invalid")
		if err := client.Connect(); err == nil {
			t.Error("Expected error with invalid key")
		}
	})

	t.Run("Connect With Invalid Host", func(t *testing.T) {
		client := NewSSHClient("nonexistent", "user", 22, "", testKey)
		if err := client.Connect(); err == nil {
			t.Error("Expected error with invalid host")
		}
	})

	t.Run("Close Without Connect", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", testKey)
		if err := client.Close(); err != nil {
			t.Error("Close should not error when not connected")
		}
	})

	t.Run("Get Client When Not Connected", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", testKey)
		if client.GetClient() != nil {
			t.Error("GetClient should return nil when not connected")
		}
	})

	t.Run("Key File Takes Precedence", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "/nonexistent/key", testKey)
		auth := client.publicKeyAuth()
		if auth != nil {
			t.Error("Expected nil auth method with invalid key file")
		}
	})
}

func TestSSHClient_KeyLoading(t *testing.T) {
	t.Run("Load Key From Data", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", testKey)
		auth := client.publicKeyAuth()
		if auth == nil {
			t.Error("Expected auth method with valid key data")
		}
	})

	t.Run("Load Key From File", func(t *testing.T) {
		// Create temp file with test key
		tmpfile, err := os.CreateTemp("", "ssh-key")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(testKey)); err != nil {
			t.Fatal(err)
		}
		tmpfile.Close()

		client := NewSSHClient("host", "user", 22, tmpfile.Name(), "")
		auth := client.publicKeyAuth()
		if auth == nil {
			t.Error("Expected auth method with valid key file")
		}
	})

	t.Run("Default Key Locations", func(t *testing.T) {
		home := t.TempDir()
		t.Setenv("HOME", home)

		// Create fake ~/.ssh/id_rsa
		sshDir := filepath.Join(home, ".ssh")
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(sshDir, "id_rsa"), []byte(testKey), 0600); err != nil {
			t.Fatal(err)
		}

		client := NewSSHClient("host", "user", 22, "", "")
		auth := client.publicKeyAuth()
		if auth == nil {
			t.Error("Expected auth method from default key location")
		}
	})

	t.Run("Invalid Key Data", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", "not-a-valid-key")
		auth := client.publicKeyAuth()
		if auth != nil {
			t.Error("Expected nil auth method with invalid key data")
		}
	})
}

func TestSSHClient_RunCommand(t *testing.T) {
	t.Run("Run Command Without Connection", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", testKey)
		_, err := client.RunCommand("echo test")
		if err == nil {
			t.Error("Expected error when running command without connection")
		}
	})
}

func TestSSHClient_ConnectionManagement(t *testing.T) {
	t.Run("Reconnection Handling", func(t *testing.T) {
		mock := NewMockSSHClient()
		
		// Set some mock outputs for basic commands
		mock.SetOutput("echo test", "test")
		
		// First connection should succeed
		if err := mock.Connect(); err != nil {
			t.Errorf("Initial connection failed: %v", err)
		}
		
		// Run a test command to verify connection
		output, err := mock.RunCommand("echo test")
		if err != nil {
			t.Errorf("Command failed: %v", err)
		}
		
		if output != "test" {
			t.Errorf("Expected output 'test', got %q", output)
		}
		
		commands := mock.GetCommands()
		if len(commands) == 0 {
			t.Error("No commands recorded")
		}
	})

	t.Run("Connection Timeout", func(t *testing.T) {
		client := NewSSHClient("slowhost", "user", 22, "", testKey)
		client.timeout = 1 * time.Millisecond // Very short timeout for testing
		
		err := client.Connect()
		if err == nil {
			t.Error("Expected timeout error")
		}
	})
}

func TestSSHClient_CommandExecution(t *testing.T) {
	t.Run("Command With Large Output", func(t *testing.T) {
		mock := NewMockSSHClient()
		
		// Generate large output
		largeOutput := strings.Repeat("x", 1024*1024) // 1MB of data
		mock.SetOutput("cat largefile", largeOutput)
		
		output, err := mock.RunCommand("cat largefile")
		if err != nil {
			t.Errorf("Failed to handle large output: %v", err)
		}
		if len(output) != len(largeOutput) {
			t.Errorf("Expected output length %d, got %d", len(largeOutput), len(output))
		}
	})

	t.Run("Command With Special Characters", func(t *testing.T) {
		mock := NewMockSSHClient()
		
		cmd := "echo 'Hello $USER' > /tmp/test; grep \"pattern\" file"
		expectedOutput := "Hello user"
		mock.SetOutput(cmd, expectedOutput)
		
		output, err := mock.RunCommand(cmd)
		if err != nil {
			t.Errorf("Failed to handle special characters: %v", err)
		}
		if output != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, output)
		}
	})
}

func TestSSHClient_FileOperations(t *testing.T) {
	t.Run("Write and Read File", func(t *testing.T) {
		mock := NewMockSSHClient()
		
		testData := []byte("test content")
		testPath := "/tmp/testfile"
		
		mock.SetOutput("read "+testPath, string(testData))
		
		err := mock.WriteFile(testPath, testData, 0644)
		if err != nil {
			t.Errorf("WriteFile failed: %v", err)
		}
		
		readData, err := mock.ReadFile(testPath)
		if err != nil {
			t.Errorf("ReadFile failed: %v", err)
		}
		
		if !bytes.Equal(testData, readData) {
			t.Error("Read data doesn't match written data")
		}
	})

	t.Run("File Permission Handling", func(t *testing.T) {
		mock := NewMockSSHClient()
		
		permissions := []uint32{0644, 0755, 0600, 0777}
		for _, perm := range permissions {
			err := mock.WriteFile("/tmp/test", []byte("test"), perm)
			if err != nil {
				t.Errorf("Failed to set permission %o: %v", perm, err)
			}
		}
		
		commands := mock.GetCommands()
		if len(commands) != len(permissions) {
			t.Errorf("Expected %d commands, got %d", len(permissions), len(commands))
		}
	})
}

func TestSSHClient_ErrorHandling(t *testing.T) {
	t.Run("Invalid Port", func(t *testing.T) {
		client := NewSSHClient("host", "user", -1, "", testKey)
		err := client.Connect()
		if err == nil {
			t.Error("Expected error with invalid port")
		}
	})

	t.Run("Empty Host", func(t *testing.T) {
		client := NewSSHClient("", "user", 22, "", testKey)
		err := client.Connect()
		if err == nil {
			t.Error("Expected error with empty host")
		}
	})

	t.Run("Invalid Key Format", func(t *testing.T) {
		client := NewSSHClient("host", "user", 22, "", "invalid-key-format")
		err := client.Connect()
		if err == nil {
			t.Error("Expected error with invalid key format")
		}
	})
}

func TestSSHClient_Validation(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		user     string
		port     int
		keyFile  string
		keyData  string
		wantErr  bool
		errMsg   string
	}{
		{
			name:    "valid configuration",
			host:    "example.com",
			user:    "user",
			port:    22,
			keyData: testKey,
			wantErr: false,
		},
		{
			name:    "empty host",
			user:    "user",
			port:    22,
			keyData: testKey,
			wantErr: true,
			errMsg:  "host cannot be empty",
		},
		{
			name:    "invalid port",
			host:    "example.com",
			user:    "user",
			port:    -1,
			keyData: testKey,
			wantErr: true,
			errMsg:  "invalid port",
		},
		{
			name:    "no authentication method",
			host:    "example.com",
			user:    "user",
			port:    22,
			wantErr: true,
			errMsg:  "no authentication method available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewSSHClient(tt.host, tt.user, tt.port, tt.keyFile, tt.keyData)
			err := client.validate()
			
			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}

	t.Run("no authentication method", func(t *testing.T) {
		// Create temp dir for HOME that doesn't contain SSH keys
		tmpHome := t.TempDir()
		origHome := os.Getenv("HOME")
		defer func() {
			os.Setenv("HOME", origHome)
		}()
		os.Setenv("HOME", tmpHome)

		// Create empty .ssh directory to ensure no default keys
		sshDir := filepath.Join(tmpHome, ".ssh")
		if err := os.MkdirAll(sshDir, 0700); err != nil {
			t.Fatal(err)
		}

		client := NewSSHClient("example.com", "user", 22, "/nonexistent/key", "")
		err := client.validate()
		if err == nil {
			t.Error("expected error but got none")
		}
		if !strings.Contains(err.Error(), "no authentication method available") {
			t.Errorf("expected 'no authentication method available' error, got: %v", err)
		}
	})
}

func (s *SSHClient) WriteFile(path string, data []byte, perm uint32) error {
	if s.client == nil {
		return fmt.Errorf("not connected")
	}
	// Implementation here
	return nil
}

func (s *SSHClient) ReadFile(path string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	// Implementation here
	return nil, nil
}
