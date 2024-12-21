package pkg

import (
	"fmt"
	"strings"
)

type MockVolumeManager struct {
	ssh SSHClientInterface
}

func NewMockVolumeManager(ssh SSHClientInterface) VolumeManagerInterface {
	return &MockVolumeManager{ssh: ssh}
}

// Implement interface methods but skip directory creation
func (m *MockVolumeManager) EnsureVolumes(volumes []string) error {
	// Only handle named volumes, skip bind mounts
	for _, volume := range volumes {
		if !strings.Contains(volume, ":") {
			if err := m.EnsureVolume(volume); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *MockVolumeManager) EnsureVolume(name string) error {
	cmd := NewRemoteCommand(m.ssh, fmt.Sprintf("podman volume ls | grep -q %s", name))
	if err := cmd.Run(); err == nil {
		return nil
	}
	cmd = NewRemoteCommand(m.ssh, fmt.Sprintf("podman volume create %s", name))
	return cmd.Run()
}

func (m *MockVolumeManager) RemoveVolume(name string) error {
	cmd := NewRemoteCommand(m.ssh, fmt.Sprintf("podman volume rm %s", name))
	return cmd.Run()
}

func (m *MockVolumeManager) ListVolumes() ([]string, error) {
	cmd := NewRemoteCommand(m.ssh, "podman volume ls --format '{{.Name}}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
} 
