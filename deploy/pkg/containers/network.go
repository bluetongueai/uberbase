package containers

import "fmt"

func (p *ContainerManager) CreateNetwork(name string) (string, error) {
	output, err := p.executor.ExecCompose("network create " + name)
	if err != nil {
		return "", fmt.Errorf("failed to create network: %w", err)
	}
	return string(output), nil
}

func (p *ContainerManager) RemoveNetwork(name string) (string, error) {
	output, err := p.executor.ExecCompose("network rm " + name)
	if err != nil {
		return "", fmt.Errorf("failed to remove network: %w", err)
	}
	return string(output), nil
}
