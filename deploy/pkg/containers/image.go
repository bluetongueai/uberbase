package containers

import "fmt"

func (p *ContainerManager) Build(tag ContainerTag) (string, error) {
	output := ""
	for _, service := range p.Compose.Project.Services {
		if service.Build == nil {
			continue
		}
		buildOutput, err := p.executor.ExecCompose("build -f " + service.Build.Dockerfile + " --tag " + service.Image + ":" + string(tag))
		if err != nil {
			return "", fmt.Errorf("failed to build image: %w", err)
		}
		output += string(buildOutput)
	}
	return output, nil
}

func (p *ContainerManager) Pull(tag ContainerTag) (string, error) {
	output := ""
	for _, service := range p.Compose.Project.Services {
		if service.Image == "" {
			continue
		}
		pullOutput, err := p.executor.ExecCompose("pull " + service.Image + ":" + string(tag))
		if err != nil {
			return "", fmt.Errorf("failed to pull image: %w", err)
		}
		output += string(pullOutput)
	}
	return output, nil
}

func (p *ContainerManager) Push(tag ContainerTag) (string, error) {
	output := ""
	for _, service := range p.Compose.Project.Services {
		if service.Image == "" {
			continue
		}
		pushOutput, err := p.executor.ExecCompose("push " + service.Image + ":" + string(tag))
		if err != nil {
			return "", fmt.Errorf("failed to push image: %w", err)
		}
		output += string(pushOutput)
	}
	return output, nil
}

func (p *ContainerManager) CompareTags(image string, firstTag ContainerTag, secondTag ContainerTag) (bool, error) {
	firstTagHash, err := p.executor.Exec("image inspect --format '{{.Id}}' " + image + ":" + string(firstTag))
	if err != nil {
		return false, fmt.Errorf("failed to get first tag hash: %w", err)
	}
	secondTagHash, err := p.executor.Exec("image inspect --format '{{.Id}}' " + image + ":" + string(secondTag))
	if err != nil {
		return false, fmt.Errorf("failed to get second tag hash: %w", err)
	}
	return firstTagHash == secondTagHash, nil
}
