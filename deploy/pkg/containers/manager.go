package containers

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/compose-spec/compose-go/v2/types"
)

type ContainerTag string

type RegistryOptions struct {
	Registry string
	Username string
	Password string
}

type ContainerExecutor interface {
	Exec(command string) (string, error)
	ExecCompose(command string) (string, error)
}

type ContainerManager struct {
	Compose  *ComposeProject
	executor ContainerExecutor
}

func NewContainerManager(executor core.Executor, compose *ComposeProject, installPodman bool) (*ContainerManager, error) {
	var containerExecutor ContainerExecutor

	if binPath, err := executor.Exec("which podman"); err == nil {
		if composePath, err := executor.Exec("which podman-compose"); err == nil {
			binPath = strings.TrimSpace(binPath)
			composePath = strings.TrimSpace(composePath)
			containerExecutor = NewPodmanExecutor(binPath, composePath, executor)
		}
	}

	if containerExecutor == nil {
		if binPath, err := executor.Exec("which docker"); err == nil {
			if composePath, err := executor.Exec("which docker-compose"); err == nil {
				binPath = strings.TrimSpace(binPath)
				composePath = strings.TrimSpace(composePath)
				containerExecutor = NewDockerExecutor(binPath, composePath, executor)
			} else {
				if _, err := executor.Exec("docker compose --version"); err == nil {
					binPath = strings.TrimSpace(binPath)
					containerExecutor = NewDockerExecutor(binPath, binPath+" compose", executor)
				}
			}
		}
	}

	if containerExecutor == nil {
		return nil, fmt.Errorf("podman/docker not found")
	}

	manager := &ContainerManager{
		Compose:  compose,
		executor: containerExecutor,
	}

	return manager, nil
}

func (p *ContainerManager) Auth(opts RegistryOptions) (string, error) {
	return p.executor.Exec("login " + opts.Registry + " -u " + opts.Username + " -p " + opts.Password)
}

func (p *ContainerManager) Up(composeOverrideFilePath string) (string, error) {
	output, err := p.executor.ExecCompose("up -d -f " + p.Compose.FilePath + " -f " + composeOverrideFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to up: %w", err)
	}
	return string(output), nil
}

func (p *ContainerManager) Down(containers []string) (string, error) {
	output, err := p.executor.ExecCompose("down --remove-orphans " + strings.Join(containers, " "))
	if err != nil {
		return "", fmt.Errorf("failed to remove: %w", err)
	}
	return string(output), nil
}

func (p *ContainerManager) Start(service *types.ServiceConfig) (string, error) {
	output, err := p.executor.ExecCompose("up -d " + service.Name)
	if err != nil {
		return "", fmt.Errorf("failed to start: %w", err)
	}
	return string(output), nil
}

func (p *ContainerManager) Run(service *types.ServiceConfig, command []string, persistent bool) (string, error) {
	runCmd := "run --rm --name " + service.Name
	if persistent {
		runCmd += " -d"
	}
	output, err := p.executor.ExecCompose(runCmd + " " + strings.Join(command, " "))
	if err != nil {
		return "", fmt.Errorf("failed to run: %w", err)
	}
	return string(output), nil
}

func (p *ContainerManager) Exec(service *types.ServiceConfig, command []string) (string, error) {
	output, err := p.executor.ExecCompose("exec " + service.Name + " " + strings.Join(command, " "))
	if err != nil {
		return "", fmt.Errorf("failed to exec: %w", err)
	}
	return string(output), nil
}
