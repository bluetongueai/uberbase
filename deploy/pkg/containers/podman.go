package containers

import (
	"fmt"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type PodmanExecutor struct {
	binPath     string
	composePath string
	executor    core.Executor
}

func NewPodmanExecutor(binPath, composePath string, executor core.Executor) *PodmanExecutor {
	return &PodmanExecutor{
		binPath:     binPath,
		composePath: composePath,
		executor:    executor,
	}
}

func (p *PodmanExecutor) Exec(command string) (string, error) {
	invocation := fmt.Sprintf("%s %s", p.binPath, command)
	return p.executor.Exec(invocation)
}

func (p *PodmanExecutor) ExecCompose(command string) (string, error) {
	invocation := fmt.Sprintf("%s %s", p.composePath, command)
	return p.executor.Exec(invocation)
}
