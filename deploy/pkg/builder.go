package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
)

type BuildOptions struct {
	Dockerfile  string
	ContextPath string
	Tag         string
	BuildArgs   map[string]string
	Environment map[string]string
	Volumes     []string
	Target      string
	NoCache     bool
	Pull        bool
	Platform    string
}

type Builder struct {
	ssh      *SSHClient
}

func NewBuilder(ssh *SSHClient) *Builder {
	return &Builder{
		ssh:      ssh,
	}
}

func (b *Builder) Build(opts BuildOptions) error {
	// Validate required fields
	if opts.ContextPath == "" {
		return fmt.Errorf("context path is required")
	}

	if opts.Tag == "" {
		return fmt.Errorf("tag is required")
	}

	// Create build command
	buildCmd := b.createBuildCommand(opts)

	// Execute the build
	cmd := NewRemoteCommand(b.ssh, buildCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

func (b *Builder) createBuildCommand(opts BuildOptions) string {
	var cmdParts []string

	cmdParts = append(cmdParts, "podman build")

	// Add Dockerfile path if specified
	if opts.Dockerfile != "" {
		cmdParts = append(cmdParts, "-f", opts.Dockerfile)
	}

	// Add tag
	cmdParts = append(cmdParts, "-t", opts.Tag)

	// Add build args
	for key, value := range opts.BuildArgs {
		cmdParts = append(cmdParts, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add target stage if specified
	if opts.Target != "" {
		cmdParts = append(cmdParts, "--target", opts.Target)
	}

	// Add no-cache flag if specified
	if opts.NoCache {
		cmdParts = append(cmdParts, "--no-cache")
	}

	// Add pull flag if specified
	if opts.Pull {
		cmdParts = append(cmdParts, "--pull")
	}

	// Add platform if specified
	if opts.Platform != "" {
		cmdParts = append(cmdParts, "--platform", opts.Platform)
	}

	// Add context path (always last)
	cmdParts = append(cmdParts, opts.ContextPath)

	return strings.Join(cmdParts, " ")
}

func (b *Builder) ValidateContextPath(contextPath string) error {
	// Check if context path exists
	cmd := NewRemoteCommand(b.ssh, fmt.Sprintf("test -d %s", contextPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("context path does not exist or is not accessible: %s", contextPath)
	}

	// Check if Dockerfile exists in context
	dockerfile := filepath.Join(contextPath, "Dockerfile")
	cmd = NewRemoteCommand(b.ssh, fmt.Sprintf("test -f %s", dockerfile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dockerfile not found in context: %s", dockerfile)
	}

	return nil
}

func (b *Builder) ListImages() ([]string, error) {
	cmd := NewRemoteCommand(b.ssh, "podman images --format '{{.Repository}}:{{.Tag}}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	images := strings.Split(strings.TrimSpace(string(output)), "\n")
	return images, nil
}

func (b *Builder) RemoveImage(tag string) error {
	cmd := NewRemoteCommand(b.ssh, fmt.Sprintf("podman rmi %s", tag))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove image %s: %w", tag, err)
	}
	return nil
}

func (b *Builder) PruneImages() error {
	cmd := NewRemoteCommand(b.ssh, "podman image prune -f")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to prune images: %w", err)
	}
	return nil
}
