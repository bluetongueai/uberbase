package containers

import (
	"strings"

	"github.com/bluetongueai/uberbase/uberbase/pkg/core"
)

type Docker struct {
	Compose     *ComposeProject
	binPath     string
	composePath string
	executor    core.Executor
}

func NewDockerExecutor(binPath, composePath string, executor core.Executor) *Docker {
	// fix compose path for current executor
	executorHome, err := executor.Exec("echo $HOME")
	if err == nil {
		composePath = strings.Replace(composePath, "~/", strings.TrimSpace(executorHome)+"/", 1)
	}
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
