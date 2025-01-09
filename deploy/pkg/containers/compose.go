package containers

import (
	"context"

	"github.com/compose-spec/compose-go/v2/cli"
	"github.com/compose-spec/compose-go/v2/types"
)

type ComposeProject struct {
	FilePath string
	Project  *types.Project
}

func NewComposeProject(composeFilePath string, projectName string) (*ComposeProject, error) {
	ctx := context.Background()

	options, err := cli.NewProjectOptions(
		[]string{composeFilePath},
		cli.WithOsEnv,
		cli.WithDotEnv,
		cli.WithName(projectName),
	)
	if err != nil {
		return nil, err
	}

	project, err := options.LoadProject(ctx)
	if err != nil {
		return nil, err
	}

	return &ComposeProject{
		FilePath: composeFilePath,
		Project:  project,
	}, nil
}
