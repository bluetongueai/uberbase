package containers

import (
	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type Docker struct {
	Compose     *ComposeProject
	binPath     string
	composePath string
	executor    core.Executor
}

func NewDockerExecutor(binPath, composePath string, executor core.Executor) *Docker {
	return &Docker{
		binPath:     binPath,
		composePath: composePath,
		executor:    executor,
	}
}
func (d *Docker) Exec(command string) (string, error) {
	return d.executor.Exec(d.binPath + " " + command)
}

func (d *Docker) ExecCompose(command string) (string, error) {
	return d.executor.Exec(d.composePath + " " + command)
}
