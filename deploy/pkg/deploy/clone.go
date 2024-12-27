package deploy

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type GitCloner struct {
	ssh *core.SSHConnection
}

func NewGitCloner(ssh *core.SSHConnection) *GitCloner {
	core.Logger.Debug("Creating new GitCloner")
	return &GitCloner{
		ssh: ssh,
	}
}

type CloneOptions struct {
	URL         string
	Branch      string
	Destination string
	Depth       int
}

func (g *GitCloner) Clone(opts CloneOptions) error {
	core.Logger.Infof("Cloning repository: %s", opts.URL)
	if err := g.validateOptions(opts); err != nil {
		core.Logger.Errorf("Invalid clone options: %v", err)
		return err
	}

	if err := g.validateGitInstalled(); err != nil {
		core.Logger.Errorf("Git not installed: %v", err)
		return err
	}

	// Ensure destination directory exists
	if err := g.ensureDirectory(opts.Destination); err != nil {
		core.Logger.Errorf("Failed to create destination directory: %v", err)
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Build clone command with proper escaping
	cmd := g.buildCloneCommand(opts)

	// Execute clone
	_, err := g.ssh.Exec(cmd)
	if err != nil {
		core.Logger.Errorf("Failed to clone repository: %v", err)
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	core.Logger.Infof("Repository cloned successfully: %s", opts.URL)
	return nil
}

func (g *GitCloner) Update(repoPath string, branch string) error {
	// Fetch latest changes
	_, err := g.ssh.Exec(fmt.Sprintf(
		"cd %s && git -c advice.detachedHead=false fetch --porcelain origin %s",
		repoPath, branch,
	))
	if err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Reset to the fetched branch
	_, err = g.ssh.Exec(fmt.Sprintf(
		"cd %s && git -c advice.detachedHead=false reset --hard origin/%s",
		repoPath, branch,
	))
	if err != nil {
		return fmt.Errorf("git reset failed: %w", err)
	}

	// Clean untracked files
	_, err = g.ssh.Exec(fmt.Sprintf(
		"cd %s && git clean -fd",
		repoPath,
	))
	if err != nil {
		return fmt.Errorf("git clean failed: %w", err)
	}

	return nil
}

func (g *GitCloner) GetRemoteURL(repoPath string) (string, error) {
	output, err := g.ssh.Exec(fmt.Sprintf(
		"cd %s && git config --get remote.origin.url",
		repoPath,
	))
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (g *GitCloner) validateGitInstalled() error {
	_, err := g.ssh.Exec("which git")
	if err != nil {
		return fmt.Errorf("git is not installed")
	}
	return nil
}

func (g *GitCloner) ensureDirectory(path string) error {
	_, err := g.ssh.Exec(fmt.Sprintf("mkdir -p %s", path))
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %w", path, err)
	}
	return nil
}

func (g *GitCloner) IsGitRepository(path string) bool {
	_, err := g.ssh.Exec(fmt.Sprintf("test -d %s/.git", path))
	return err == nil
}

func (g *GitCloner) GetCurrentCommit(repoPath string) (string, error) {
	output, err := g.ssh.Exec(fmt.Sprintf(
		"cd %s && git rev-parse HEAD",
		repoPath,
	))
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (g *GitCloner) validateOptions(opts CloneOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("repository URL is required")
	}
	if opts.Destination == "" {
		return fmt.Errorf("destination path is required")
	}
	return nil
}

func (g *GitCloner) buildCloneCommand(opts CloneOptions) string {
	cmdParts := []string{"git", "clone"}

	if opts.Branch != "" {
		cmdParts = append(cmdParts, "-b", opts.Branch)
	}

	if opts.Depth > 0 {
		cmdParts = append(cmdParts, "--depth", fmt.Sprintf("%d", opts.Depth))
	}

	// Only quote URLs and paths if they contain special characters
	if strings.ContainsAny(opts.URL, " $&\"'") {
		cmdParts = append(cmdParts, fmt.Sprintf("%q", opts.URL))
	} else {
		cmdParts = append(cmdParts, opts.URL)
	}

	if opts.Destination != "" {
		if strings.ContainsAny(opts.Destination, " $&\"'") {
			cmdParts = append(cmdParts, fmt.Sprintf("%q", opts.Destination))
		} else {
			cmdParts = append(cmdParts, opts.Destination)
		}
	}

	return strings.Join(cmdParts, " ")
}
