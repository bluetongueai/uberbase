package containers

import "fmt"

func (p *ContainerManager) CreateVolume(name string) (string, error) {
	output, err := p.executor.ExecCompose("volume create " + name)
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}
	return string(output), nil
}

func (p *ContainerManager) RemoveVolume(name string) (string, error) {
	output, err := p.executor.ExecCompose("volume rm " + name)
	if err != nil {
		return "", fmt.Errorf("failed to remove volume: %w", err)
	}
	return string(output), nil
}
