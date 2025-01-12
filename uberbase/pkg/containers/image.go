package containers

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/uberbase/pkg/utils"
)

func (p *ContainerManager) Build(tag ContainerTag) (string, error) {
	output := ""
	for _, service := range p.Compose.Project.Services {
		if service.Build == nil {
			continue
		}
		image := utils.StripTag(service.Image)
		buildArgs := []string{}
		if service.Build.Args != nil {
			for k, v := range service.Build.Args {
				buildArgs = append(buildArgs, "--build-arg", k+"="+*v)
			}
		}
		buildArgs = append(buildArgs, "--tag", image+":"+string(tag))

		if service.Build.Dockerfile != "" {
			if !strings.HasPrefix(service.Build.Dockerfile, service.Build.Context) {
				buildArgs = append(buildArgs, "-f", service.Build.Context+"/"+service.Build.Dockerfile)
			} else {
				buildArgs = append(buildArgs, "-f", service.Build.Dockerfile)
			}
		}

		buildArgs = append(buildArgs, service.Build.Context)
		buildOutput, err := p.executor.Exec("builder build " + strings.Join(buildArgs, " "))
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
		image := service.Image
		if service.Build != nil {
			image = utils.StripTag(service.Image) + ":" + string(tag)
		}
		pullOutput, err := p.executor.Exec("pull " + image)
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
		if service.Build == nil {
			continue
		}
		image := utils.StripTag(service.Image)
		pushOutput, err := p.executor.Exec("push " + image + ":" + string(tag))
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
