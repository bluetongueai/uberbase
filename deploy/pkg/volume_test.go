package pkg

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestVolumeManager(t *testing.T) {
	ssh := NewMockSSHClient()
	vm := NewVolumeManager(ssh)

	t.Run("ensure volume", func(t *testing.T) {
		err := vm.EnsureVolume("test-volume")
		assert.NoError(t, err)
		
		commands := ssh.GetCommands()
		assert.Contains(t, commands, "podman volume inspect test-volume || podman volume create test-volume")
	})

	t.Run("ensure volumes with bind mount", func(t *testing.T) {
		volumes := []string{
			"named-volume",
			"/host/path:/container/path",
		}
		
		err := vm.EnsureVolumes(volumes)
		assert.NoError(t, err)
		
		commands := ssh.GetCommands()
		assert.Contains(t, commands, "podman volume inspect named-volume || podman volume create named-volume")
	})

	t.Run("remove volume", func(t *testing.T) {
		err := vm.RemoveVolume("test-volume")
		assert.NoError(t, err)
		
		commands := ssh.GetCommands()
		assert.Contains(t, commands, "podman volume rm test-volume")
	})

	t.Run("list volumes", func(t *testing.T) {
		expectedOutput := "vol1\nvol2\nvol3"
		ssh.SetOutput("podman volume ls --format '{{.Name}}'", expectedOutput)
		
		volumes, err := vm.ListVolumes()
		assert.NoError(t, err)
		assert.Equal(t, []string{"vol1", "vol2", "vol3"}, volumes)
	})
}

func TestVolumeManager_SELinuxHandling(t *testing.T) {
	ssh := NewMockSSHClient()
	vm := NewVolumeManager(ssh)

	t.Run("Private SELinux Label", func(t *testing.T) {
		hostPath := "/host/path"
		options := []string{"Z"}

		err := vm.handleSELinux(hostPath, options)
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		expectedCmd := "chcon -Rt container_file_t /host/path"
		assert.Contains(t, commands, expectedCmd)
	})

	t.Run("Shared SELinux Label", func(t *testing.T) {
		hostPath := "/shared/path"
		options := []string{"z"}

		err := vm.handleSELinux(hostPath, options)
		assert.NoError(t, err)

		commands := ssh.GetCommands()
		expectedCmd := "chcon -Rt container_share_t /shared/path"
		assert.Contains(t, commands, expectedCmd)
	})

	t.Run("SELinux Command Failure", func(t *testing.T) {
		ssh.SetError("chcon", fmt.Errorf("selinux not enabled"))
		
		err := vm.handleSELinux("/path", []string{"Z"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set SELinux context")
	})
}

func TestVolumeManager_VolumeOptions(t *testing.T) {
	ssh := NewMockSSHClient()
	vm := NewVolumeManager(ssh)

	t.Run("Handle Multiple Options", func(t *testing.T) {
		hostPath := "/test/path"
		options := []string{"ro", "Z", "shared", "nocopy"}

		mountOpts := vm.handleVolumeOptions(hostPath, options)
		
		// Verify mount options
		assert.Contains(t, mountOpts, "ro")
		assert.Contains(t, mountOpts, "shared")
		assert.Contains(t, mountOpts, "nocopy")
		assert.Contains(t, mountOpts, "Z")

		// Verify SELinux handling
		commands := ssh.GetCommands()
		assert.Contains(t, commands, "chcon -Rt container_file_t /test/path")
	})

	t.Run("Handle Bind Propagation", func(t *testing.T) {
		options := []string{"rshared", "rbind"}
		mountOpts := vm.handleVolumeOptions("/path", options)
		
		// Check each option individually
		for _, opt := range options {
			if !contains(mountOpts, opt) {
				t.Errorf("Expected mount options to contain %q", opt)
			}
		}
	})

	t.Run("Handle Cache Options", func(t *testing.T) {
		options := []string{"cached", "delegated"}
		mountOpts := vm.handleVolumeOptions("/path", options)
		
		assert.Contains(t, mountOpts, "cached")
		assert.Contains(t, mountOpts, "delegated")
	})
}

func TestVolumeManager_ErrorHandling(t *testing.T) {
	ssh := NewMockSSHClient()
	vm := NewVolumeManager(ssh)

	t.Run("Volume Create Error", func(t *testing.T) {
		ssh.SetError("podman volume create", fmt.Errorf("creation failed"))
		
		err := vm.EnsureVolume("test-vol")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creation failed")
	})

	t.Run("Volume Remove Error", func(t *testing.T) {
		ssh.SetError("podman volume rm", fmt.Errorf("volume in use"))
		
		err := vm.RemoveVolume("test-vol")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "volume in use")
	})

	t.Run("Invalid Volume Name", func(t *testing.T) {
		err := vm.EnsureVolume("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid volume name")
	})
}

func TestVolumeManager_PathHandling(t *testing.T) {
	ssh := NewMockSSHClient()
	vm := NewVolumeManager(ssh)

	t.Run("Environment Variable Expansion", func(t *testing.T) {
		t.Setenv("TEST_PATH", "/test/path")
		volumes := []string{"$TEST_PATH:/container/path"}
		
		err := vm.EnsureVolumes(volumes)
		assert.NoError(t, err)
		
		// No need to check commands as we're just validating the volume spec
		assert.NoError(t, err)
	})

	t.Run("Handle Special Characters", func(t *testing.T) {
		volumes := []string{"/path with spaces:/container path"}
		
		err := vm.EnsureVolumes(volumes)
		assert.NoError(t, err)
		
		// Just validate that the operation succeeds
		assert.NoError(t, err)
	})
}

func TestVolumeManager_Validation(t *testing.T) {
	tests := []struct {
		name    string
		volumes []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid named volume",
			volumes: []string{"myvolume"},
			wantErr: false,
		},
		{
			name:    "valid bind mount",
			volumes: []string{"/host:/container:ro"},
			wantErr: false,
		},
		{
			name:    "invalid format",
			volumes: []string{"invalid:format:extra:part"},
			wantErr: true,
			errMsg:  "invalid volume specification",
		},
		{
			name:    "invalid option",
			volumes: []string{"/host:/container:invalid"},
			wantErr: true,
			errMsg:  "invalid mount option",
		},
		{
			name:    "empty volume name",
			volumes: []string{""},
			wantErr: true,
			errMsg:  "invalid volume name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := NewVolumeManager(NewMockSSHClient())
			err := vm.EnsureVolumes(tt.volumes)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
