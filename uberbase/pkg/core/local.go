package core

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/bluetongueai/uberbase/uberbase/pkg/logging"
)

type LocalExecutor struct {
	installer *Installer
}

func NewLocalExecutor() *LocalExecutor {
	executor := &LocalExecutor{}
	executor.installer = NewInstaller(executor)
	return executor
}

func (e *LocalExecutor) Test() bool {
	return true
}

func (e *LocalExecutor) SendFile(localPath, remotePath string) error {
	return nil
}

func (e *LocalExecutor) Exec(command string) (string, error) {
	logging.Logger.Infof("local: \033[33m%s\033[0m", command)

	cmd := exec.Command("sh", "-c", command)

	// Create a buffer for stderr
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Get stdout
	output, err := cmd.Output()

	// If there's stderr output, log it in red regardless of error
	if stderrStr := stderr.String(); stderrStr != "" && err != nil {
		logging.Logger.Infof("\033[31m%s\033[0m", stderrStr)
	}

	if err != nil {
		return "", err
	}

	return string(output), nil
}

func (e *LocalExecutor) Verify() error {
	if !e.installer.HasGit() {
		return fmt.Errorf("git is not installed")
	}
	if !e.installer.HasDocker() {
		if !e.installer.HasPodman() {
			return fmt.Errorf("docker or podman is not installed")
		}
	}
	if !e.installer.HasDockerCompose() {
		if !e.installer.HasPodmanCompose() {
			return fmt.Errorf("docker-compose or podman-compose is not installed")
		}
	}
	return nil
}
