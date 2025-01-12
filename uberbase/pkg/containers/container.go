package containers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
)

type ContainerState struct {
	Status     string `json:"status"`
	Running    bool   `json:"running"`
	Paused     bool   `json:"paused"`
	Restarting bool   `json:"restarting"`
	OOMKilled  bool   `json:"oom_killed"`
	Dead       bool   `json:"dead"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	Pid        int    `json:"pid"`
	ExitCode   int    `json:"exit_code"`
	Error      string `json:"error"`
}

type ContainerInspectInfo struct {
	ID    string         `json:"id"`
	State ContainerState `json:"state"`
}

func (p *ContainerManager) GetContainerTag(service *types.ServiceConfig) (ContainerTag, error) {
	container, err := p.executor.ExecCompose("inspect -f '{{.Config.Image}}' " + service.Name)
	if err != nil {
		return "", fmt.Errorf("failed to get container tag: %w", err)
	}
	parts := strings.Split(string(container), ":")
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to get container tag: %w", err)
	}
	return ContainerTag(parts[1]), nil
}

func (p *ContainerManager) Inspect(containerID string) (ContainerInspectInfo, error) {
	var inspectInfo ContainerInspectInfo
	output, err := p.executor.ExecCompose("inspect " + containerID)
	if err != nil {
		return ContainerInspectInfo{}, fmt.Errorf("failed to inspect container: %w", err)
	}
	err = json.Unmarshal([]byte(output), &inspectInfo)
	if err != nil {
		return ContainerInspectInfo{}, fmt.Errorf("failed to unmarshal inspect output: %w", err)
	}
	return inspectInfo, nil
}
