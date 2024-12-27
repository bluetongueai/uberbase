package deploy

import (
	"fmt"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg"
	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type BackupManager struct {
	ssh       *core.SSHConnection
	backupDir string
	retention time.Duration
}

func NewBackupManager(ssh *core.SSHConnection) *BackupManager {
	return &BackupManager{
		ssh:       ssh,
		backupDir: "/var/lib/podman/volumes/backups",
		retention: 7 * 24 * time.Hour, // 1 week default retention
	}
}

func (b *BackupManager) CleanupOldBackups(service *Service) error {
	// Find backups older than retention period
	cmd := fmt.Sprintf(`find %s -name "%s-*" -type f -mtime +%d -delete`,
		b.backupDir, service.Name, int(b.retention.Hours()/24))

	if _, err := b.ssh.Exec(cmd); err != nil {
		return fmt.Errorf("failed to cleanup old backups: %w", err)
	}

	return nil
}

func (b *BackupManager) ValidateBackupSpace(service *Service) error {
	// Check available space in backup directory
	cmd := fmt.Sprintf("df --output=avail %s | tail -n 1", b.backupDir)
	output, err := b.ssh.Exec(cmd)
	if err != nil {
		return fmt.Errorf("failed to check backup space: %w", err)
	}

	availableKB := int64(pkg.ParseInt(string(output)))
	requiredKB := int64(1024 * 1024) // 1GB minimum

	if availableKB < requiredKB {
		return fmt.Errorf("insufficient space for backup: %d KB available, %d KB required",
			availableKB, requiredKB)
	}

	return nil
}
