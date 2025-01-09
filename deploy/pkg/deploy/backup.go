package deploy

import (
	"fmt"
	"os"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/containers"
	"github.com/compose-spec/compose-go/v2/types"
)

type BackupManager struct {
	containerMgr *containers.ContainerManager
	retention    time.Duration
}

func NewBackupManager(containerMgr *containers.ContainerManager) *BackupManager {
	return &BackupManager{
		containerMgr: containerMgr,
		retention:    7 * 24 * time.Hour, // 1 week default retention
	}
}

func (b *BackupManager) BackupVolume(volumeName string) error {
	backupDir := fmt.Sprintf("/var/lib/podman/volumes/backups/%s", volumeName)
	backupFile := fmt.Sprintf("%s-%s.tar.gz", volumeName, time.Now().Format("20250101-120000"))

	if err := b.reapBackups(); err != nil {
		return fmt.Errorf("failed to reap old backups: %w", err)
	}

	// use the container manager to start an alpine container with the volume mounted
	// and run the tar command to create the backup
	_, err := b.containerMgr.Run(&types.ServiceConfig{
		Name:  "backup-" + volumeName,
		Image: "alpine",
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   "volume",
				Source: volumeName,
				Target: "/data",
			},
			{
				Type:   "bind",
				Source: backupDir,
				Target: "/backups",
			},
		},
	}, []string{"sh", "-c", fmt.Sprintf("tar -czvf /backups/%s.tar.gz /data", backupFile)}, false)

	if err != nil {
		return fmt.Errorf("failed to create backup container: %w", err)
	}

	return nil
}

func (b *BackupManager) RestoreVolume(volumeName, backupFile string) error {
	// use the container manager to start an alpine container with the volume mounted
	// and run the tar command to create the backup
	_, err := b.containerMgr.Run(&types.ServiceConfig{
		Name:  "restore-" + volumeName,
		Image: "alpine",
		Volumes: []types.ServiceVolumeConfig{
			{
				Type:   "bind",
				Source: "/var/lib/podman/volumes/backups",
				Target: "/backups",
			},
			{
				Type:   "volume",
				Source: volumeName,
				Target: "/data",
			},
		},
	}, []string{"sh", "-c", fmt.Sprintf("tar -xzvf /backups/%s.tar.gz -C /data", backupFile)}, false)

	if err != nil {
		return fmt.Errorf("failed to restore backup: %w", err)
	}

	return nil
}

func (b *BackupManager) reapBackups() error {
	backups, err := os.ReadDir("/var/lib/podman/volumes/backups")
	if err != nil {
		return fmt.Errorf("failed to read backups directory: %w", err)
	}

	for _, backup := range backups {
		backupTime, err := time.Parse("20060102-150405", backup.Name())
		if err != nil {
			return fmt.Errorf("failed to parse backup time: %w", err)
		}

		if backupTime.Before(time.Now().Add(-b.retention)) {
			os.RemoveAll(backup.Name())
		}
	}

	return nil
}
