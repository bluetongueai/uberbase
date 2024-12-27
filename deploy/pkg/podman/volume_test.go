package podman

import (
	"testing"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/stretchr/testify/assert"
)

func TestVolumeManager(t *testing.T) {
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

	vm := NewVolumeManager(conn)

	t.Run("ensure volume", func(t *testing.T) {
		mock.SetReturnString("")
		err := vm.EnsureVolume("test-volume")
		assert.NoError(t, err)
	})

	t.Run("ensure volumes with bind mount", func(t *testing.T) {
		mock.SetReturnString("")
		volumes := []string{
			"named-volume",
			"/host/path:/container/path",
		}

		err := vm.EnsureVolumes(volumes)
		assert.NoError(t, err)
	})

	t.Run("remove volume", func(t *testing.T) {
		mock.SetReturnString("")
		err := vm.RemoveVolume("test-volume")
		assert.NoError(t, err)
	})

	t.Run("list volumes", func(t *testing.T) {
		expectedOutput := "vol1\nvol2\nvol3"
		mock.SetReturnString(expectedOutput)

		volumes, err := vm.ListVolumes()
		assert.NoError(t, err)
		assert.Equal(t, []string{"vol1", "vol2", "vol3"}, volumes)
	})
}

func TestVolumeManager_SELinuxHandling(t *testing.T) {
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

	vm := NewVolumeManager(conn)

	t.Run("Private SELinux Label", func(t *testing.T) {
		mock.SetReturnString("")
		hostPath := "/host/path"
		options := []string{"Z"}

		err := vm.handleSELinux(hostPath, options)
		assert.NoError(t, err)
	})

	t.Run("Shared SELinux Label", func(t *testing.T) {
		mock.SetReturnString("")
		hostPath := "/shared/path"
		options := []string{"z"}

		err := vm.handleSELinux(hostPath, options)
		assert.NoError(t, err)
	})

	t.Run("SELinux Command Failure", func(t *testing.T) {
		mock.SetReturnString("selinux not enabled")
		err := vm.handleSELinux("/path", []string{"Z"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to set SELinux context")
	})
}

func TestVolumeManager_VolumeOptions(t *testing.T) {
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

	vm := NewVolumeManager(conn)

	t.Run("Handle Multiple Options", func(t *testing.T) {
		hostPath := "/test/path"
		options := []string{"ro", "Z", "shared", "nocopy"}

		mountOpts := vm.handleVolumeOptions(hostPath, options)

		assert.Contains(t, mountOpts, "ro")
		assert.Contains(t, mountOpts, "shared")
		assert.Contains(t, mountOpts, "nocopy")
		assert.Contains(t, mountOpts, "Z")
	})
}

func TestVolumeManager_ErrorHandling(t *testing.T) {
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

	vm := NewVolumeManager(conn)

	t.Run("Volume Create Error", func(t *testing.T) {
		mock.SetReturnString("creation failed")

		err := vm.EnsureVolume("test-vol")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "creation failed")
	})

	t.Run("Volume Remove Error", func(t *testing.T) {
		mock.SetReturnString("volume in use")

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

func TestVolumeManager_Validation(t *testing.T) {
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
			vm := NewVolumeManager(conn)
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
