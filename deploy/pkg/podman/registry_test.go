package podman

import (
	"testing"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestRegistryClient(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	t.Run("Push Image", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{
			Name: "myapp",
			Tag:  "v1.0",
		}

		mock.SetReturnString("")

		err = client.PushImage(imageRef)
		assert.NoError(t, err)
	})

	t.Run("Pull Image", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{
			Name: "myapp",
			Tag:  "latest",
		}

		mock.SetReturnString("")

		err = client.PullImage(imageRef)
		assert.NoError(t, err)
	})

	t.Run("Tag Image", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		sourceRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		targetRef := ImageRef{Name: "myapp", Tag: "latest"}

		mock.SetReturnString("")

		err = client.TagImage(sourceRef, targetRef)
		assert.NoError(t, err)
	})

	t.Run("Image Exists", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}

		mock.SetReturnString("")
		exists, err := client.ImageExists(imageRef)
		assert.NoError(t, err)
		assert.True(t, exists)

		mock.SetReturnString("not found")
		exists, err = client.ImageExists(imageRef)
		assert.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("List Tags", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{Name: "myapp"}
		expectedTags := "v1.0\nv1.1\nlatest"
		mock.SetReturnString(expectedTags)

		tags, err := client.ListTags(imageRef)
		assert.NoError(t, err)
		assert.Equal(t, []string{"v1.0", "v1.1", "latest"}, tags)
	})

	t.Run("Delete Image", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		mock.SetReturnString("")

		err = client.DeleteImage(imageRef)
		assert.NoError(t, err)
	})

	t.Run("Get Image Digest", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		expectedDigest := "sha256:abc123"
		mock.SetReturnString(expectedDigest)

		digest, err := client.GetImageDigest(imageRef)
		assert.NoError(t, err)
		assert.Equal(t, expectedDigest, digest)
	})

	t.Run("Login Error", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		mock.SetReturnString("login failed")

		imageRef := ImageRef{Name: "myapp", Tag: "v1.0"}
		err = client.PushImage(imageRef)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
	})

	t.Run("Network Error", func(t *testing.T) {
		ssh, err := core.NewSession(core.SSHConfig{
			Host:    "localhost:2222",
			User:    "user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)
		client := New(ssh, RegistryConfig{
			Host:     "registry.example.com",
			Username: "user",
			Password: "pass",
		})

		mock.SetReturnString("network timeout")

		err = client.PushImage(ImageRef{Name: "myapp", Tag: "v1.0"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "login failed")
	})
}

func TestRegistryClient_Validation(t *testing.T) {
	mock := core.NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	ssh, err := core.NewSession(core.SSHConfig{
		Host:    "localhost:2222",
		User:    "user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	client := New(ssh, RegistryConfig{
		Host:     "registry.example.com",
		Username: "user",
		Password: "pass",
	})

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
				Name: "myapp",
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
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
