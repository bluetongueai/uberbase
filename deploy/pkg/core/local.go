package core

import (
	"fmt"
	"os/exec"
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

func (e *LocalExecutor) Exec(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
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
