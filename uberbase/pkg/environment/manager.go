package environment

import (
	"fmt"
	"os"
	"strings"

	"github.com/bluetongueai/uberbase/uberbase/pkg/containers"
	"github.com/bluetongueai/uberbase/uberbase/pkg/core"
)

type EnvironmentManager struct {
	remoteExecutor *core.RemoteExecutor
}

func NewEnvironmentManager(remoteExecutor *core.RemoteExecutor) *EnvironmentManager {
	return &EnvironmentManager{
		remoteExecutor: remoteExecutor,
	}
}

func (e *EnvironmentManager) DeployEnv(tag containers.ContainerTag) error {
	if err := e.writeRemoteEnv(string(tag)); err != nil {
		return fmt.Errorf("failed to write remote env: %w", err)
	}

	return nil
}

func (e *EnvironmentManager) CommitEnv(tag containers.ContainerTag) error {
	if err := e.deleteRemoteEnv(string(tag)); err != nil {
		return fmt.Errorf("failed to delete remote env: %w", err)
	}

	return nil
}

func (e *EnvironmentManager) writeRemoteEnv(path string) error {
	prefixes := os.Getenv("UBERBASE_DEPLOY_PREFIXES")

	envVars := []string{}
	for _, prefix := range strings.Split(prefixes, ",") {
		for _, envVar := range os.Environ() {
			if strings.HasPrefix(envVar, string(prefix)) {
				envVars = append(envVars, envVar)
			}
		}
	}

	envFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}
	defer envFile.Close()

	envFile.WriteString(strings.Join(envVars, "\n"))

	return nil
}

func (e *EnvironmentManager) deleteRemoteEnv(path string) error {
	return os.Remove(path)
}
