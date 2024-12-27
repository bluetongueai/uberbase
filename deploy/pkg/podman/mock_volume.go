package podman

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type MockVolumeManager struct {
	ssh *core.SSHConnection
}

func NewMockVolumeManager(ssh *core.SSHConnection) *MockVolumeManager {
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
	_, err := m.ssh.Exec(fmt.Sprintf("podman volume ls | grep -q %s", name))
	if err == nil {
		return nil
	}
	_, err = m.ssh.Exec(fmt.Sprintf("podman volume create %s", name))
	return err
}

func (m *MockVolumeManager) RemoveVolume(name string) error {
	_, err := m.ssh.Exec(fmt.Sprintf("podman volume rm %s", name))
	return err
}

func (m *MockVolumeManager) ListVolumes() ([]string, error) {
	output, err := m.ssh.Exec("podman volume ls --format '{{.Name}}'")
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(output)), "\n"), nil
}
