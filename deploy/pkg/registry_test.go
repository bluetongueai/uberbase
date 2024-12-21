package pkg

import (
	"encoding/json"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryClient(t *testing.T) {
	t.Run("Push Image", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", &RegistryAuth{
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{
			Name: "myapp",
			Tag:  "v1.0",
		}

		mock.SetOutput("podman login -u user -p pass registry.example.com", "")
		mock.SetOutput("podman push registry.example.com/myapp", "")

		if err := client.PushImage(imageRef); err != nil {
			t.Errorf("PushImage failed: %v", err)
		}

		commands := mock.GetCommands()
		if len(commands) < 2 {
			t.Fatalf("Expected at least 2 commands, got %d", len(commands))
		}
		assert.Contains(t, commands[0], "podman login")
		assert.Contains(t, commands[1], "podman push")
	})

	t.Run("Pull Image", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		imageRef := ImageRef{
			Name: "myapp",
			Tag:  "latest",
		}

		mock.SetOutput("podman pull", "")

		if err := client.PullImage(imageRef); err != nil {
			t.Errorf("PullImage failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman pull registry.example.com/myapp:latest"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("Tag Image", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		sourceRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		targetRef := ImageRef{Name: "myapp", Tag: "latest"}

		mock.SetOutput("podman tag", "")

		if err := client.TagImage(sourceRef, targetRef); err != nil {
			t.Errorf("TagImage failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman tag registry.example.com/myapp:v1.0 registry.example.com/myapp:latest"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("Image Exists", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}

		mock.SetOutput("podman inspect", "")

		exists, err := client.ImageExists(imageRef)
		if err != nil {
			t.Errorf("ImageExists failed: %v", err)
		}
		if !exists {
			t.Error("Expected image to exist")
		}

		mock.SetError("podman inspect", fmt.Errorf("not found"))
		exists, err = client.ImageExists(imageRef)
		if err != nil {
			t.Errorf("ImageExists failed: %v", err)
		}
		if exists {
			t.Error("Expected image to not exist")
		}
	})

	t.Run("List Tags", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		imageRef := ImageRef{Name: "myapp"}
		expectedTags := "v1.0\nv1.1\nlatest"
		mock.SetOutput("podman search --list-tags", expectedTags)

		tags, err := client.ListTags(imageRef)
		if err != nil {
			t.Errorf("ListTags failed: %v", err)
		}

		if len(tags) != 3 {
			t.Errorf("Expected 3 tags, got %d", len(tags))
		}

		expectedTagList := []string{"v1.0", "v1.1", "latest"}
		for i, tag := range tags {
			if tag != expectedTagList[i] {
				t.Errorf("Expected tag %q, got %q", expectedTagList[i], tag)
			}
		}
	})

	t.Run("Delete Image", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		mock.SetOutput("podman rmi", "")

		if err := client.DeleteImage(imageRef); err != nil {
			t.Errorf("DeleteImage failed: %v", err)
		}

		commands := mock.GetCommands()
		expectedCmd := "podman rmi registry.example.com/myapp:v1.0"
		if commands[0] != expectedCmd {
			t.Errorf("Expected command %q, got %q", expectedCmd, commands[0])
		}
	})

	t.Run("Get Image Digest", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		expectedDigest := "sha256:abc123"
		mock.SetOutput("podman inspect", expectedDigest)

		digest, err := client.GetImageDigest(imageRef)
		if err != nil {
			t.Errorf("GetImageDigest failed: %v", err)
		}
		if digest != expectedDigest {
			t.Errorf("Expected digest %q, got %q", expectedDigest, digest)
		}
	})

	t.Run("Login Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", &RegistryAuth{
			Username: "user",
			Password: "pass",
		})

		mock.SetError("podman login -u user -p pass registry.example.com", fmt.Errorf("login failed"))

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		err := client.PushImage(imageRef)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
	})

	t.Run("Network Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", &RegistryAuth{
			Username: "user",
			Password: "pass",
		})

		mock.SetError("podman login -u user -p pass registry.example.com", fmt.Errorf("network timeout"))
		
		err := client.PushImage(ImageRef{Name: "myapp", Tag: "v1.0"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
	})
}

func TestRegistryClient_Authentication(t *testing.T) {
	t.Run("Auth Token Generation", func(t *testing.T) {
			auth := &RegistryAuth{
				Username: "testuser",
				Password: "testpass",
			}
			client := NewRegistryClient(NewMockSSHClient(), "registry.example.com", auth)

			// Verify auth token format
			if client.authToken == "" {
				t.Error("Expected auth token to be generated")
			}
			decoded, err := base64.StdEncoding.DecodeString(client.authToken)
			if err != nil {
				t.Errorf("Invalid base64 encoding: %v", err)
			}

			var authData map[string]string
			if err := json.Unmarshal(decoded, &authData); err != nil {
				t.Errorf("Invalid JSON in auth token: %v", err)
			}

			if authData["username"] != auth.Username || authData["password"] != auth.Password {
				t.Error("Auth token contains incorrect credentials")
			}
		})

	t.Run("Skip Login Without Credentials", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		err := client.PushImage(ImageRef{Name: "myapp", Tag: "v1.0"})
		if err != nil {
			t.Errorf("Push failed: %v", err)
		}

		commands := mock.GetCommands()
		for _, cmd := range commands {
			if strings.Contains(cmd, "podman login") {
				t.Error("Login command should not be executed without credentials")
			}
		}
	})
}

func TestRegistryClient_ImageReferenceHandling(t *testing.T) {
	mock := NewMockSSHClient()
	client := NewRegistryClient(mock, "registry.example.com", nil)

	tests := []struct {
		name     string
		ref      ImageRef
		expected string
	}{
		{
			name: "full reference",
			ref: ImageRef{
				Registry: "custom.registry.com",
				Name:     "myapp",
				Tag:      "v1.0",
			},
			expected: "custom.registry.com/myapp:v1.0",
		},
		{
			name: "default registry",
			ref: ImageRef{
				Name: "myapp",
				Tag:  "v1.0",
			},
			expected: "registry.example.com/myapp:v1.0",
		},
		{
			name: "default tag",
			ref: ImageRef{
				Name: "myapp",
			},
			expected: "registry.example.com/myapp:latest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.getFullImageRef(tt.ref)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestRegistryClient_ErrorHandling(t *testing.T) {
	mock := NewMockSSHClient()
	client := NewRegistryClient(mock, "registry.example.com", &RegistryAuth{
		Username: "user",
		Password: "pass",
	})

	t.Run("Network Error", func(t *testing.T) {
		mock.SetError("podman login", fmt.Errorf("network timeout"))
		
		err := client.PushImage(ImageRef{Name: "myapp", Tag: "v1.0"})
		if err == nil {
			t.Error("Expected error on network timeout")
		}
		if !strings.Contains(err.Error(), "login failed") {
			t.Errorf("Expected login failure error, got: %v", err)
		}
	})

	t.Run("Invalid Image Reference", func(t *testing.T) {
		mock.SetOutput("podman pull", "")
		
		err := client.PullImage(ImageRef{Name: ""})
		if err == nil {
			t.Error("Expected error for invalid image reference")
		}
	})

	t.Run("Image Not Found", func(t *testing.T) {
		mock := NewMockSSHClient()
		client := NewRegistryClient(mock, "registry.example.com", nil)

		// Mock the login command to succeed
		mock.SetOutput("podman login", "Login Succeeded!")
		
		// Mock the pull command to fail with "not found" error
		mock.SetError("podman pull", fmt.Errorf("error: manifest unknown: manifest unknown"))

		err := client.PullImage(ImageRef{
			Name: "nonexistent",
			Tag:  "latest",
		})
		if err == nil {
			t.Error("Expected pull failure error, got none")
		}
		if !strings.Contains(err.Error(), "manifest unknown") {
			t.Errorf("Expected 'manifest unknown' error, got: %v", err)
		}
	})
}

func TestRegistryClient_ConcurrentOperations(t *testing.T) {
	mock := NewMockSSHClient()
	client := NewRegistryClient(mock, "registry.example.com", nil)

	t.Run("Concurrent Pulls", func(t *testing.T) {
		mock.SetOutput("podman pull", "")

		var wg sync.WaitGroup
		errChan := make(chan error, 5)

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				ref := ImageRef{
					Name: fmt.Sprintf("app%d", i),
					Tag:  "latest",
				}
				if err := client.PullImage(ref); err != nil {
					errChan <- fmt.Errorf("pull %d failed: %v", i, err)
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		for err := range errChan {
			t.Errorf("Concurrent pull error: %v", err)
		}
	})
}

func TestRegistryClient_Validation(t *testing.T) {
	client := NewRegistryClient(NewMockSSHClient(), "registry.example.com", nil)

	tests := []struct {
		name    string
		ref     ImageRef
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty name",
			ref: ImageRef{
				Tag: "latest",
			},
			wantErr: true,
			errMsg:  "image name is required",
		},
		{
			name: "invalid registry URL",
			ref: ImageRef{
				Registry: "invalid url",
				Name:     "myapp",
			},
			wantErr: true,
			errMsg:  "invalid registry URL",
		},
		{
			name: "valid reference",
			ref: ImageRef{
				Name: "myapp",
				Tag:  "v1.0",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := client.validateImageRef(tt.ref)
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
