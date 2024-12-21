package pkg

import (
	"fmt"
	"strings"
)

type GitCloner struct {
	ssh SSHClientInterface
}

func NewGitCloner(ssh SSHClientInterface) *GitCloner {
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
	if err := g.validateOptions(opts); err != nil {
		return err
	}

	if err := g.validateGitInstalled(); err != nil {
		return err
	}

	// Ensure destination directory exists
	if err := g.ensureDirectory(opts.Destination); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Build clone command with proper escaping
	cmd := g.buildCloneCommand(opts)
	
	// Execute clone
	execCmd := NewRemoteCommand(g.ssh, cmd)
	if err := execCmd.Run(); err != nil {
		return fmt.Errorf("git clone failed: %w", err)
	}

	return nil
}

func (g *GitCloner) Update(repoPath string, branch string) error {
	// Fetch latest changes
	fetchCmd := NewRemoteCommand(g.ssh, fmt.Sprintf(
		"cd %s && git -c advice.detachedHead=false fetch --porcelain origin %s",
		repoPath, branch,
	))
	if err := fetchCmd.Run(); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}

	// Reset to the fetched branch
	resetCmd := NewRemoteCommand(g.ssh, fmt.Sprintf(
		"cd %s && git -c advice.detachedHead=false reset --hard origin/%s",
		repoPath, branch,
	))
	if err := resetCmd.Run(); err != nil {
		return fmt.Errorf("git reset failed: %w", err)
	}

	// Clean untracked files
	cleanCmd := NewRemoteCommand(g.ssh, fmt.Sprintf(
		"cd %s && git clean -fd",
		repoPath,
	))
	if err := cleanCmd.Run(); err != nil {
		return fmt.Errorf("git clean failed: %w", err)
	}

	return nil
}

func (g *GitCloner) GetRemoteURL(repoPath string) (string, error) {
	cmd := NewRemoteCommand(g.ssh, fmt.Sprintf(
		"cd %s && git config --get remote.origin.url",
		repoPath,
	))
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (g *GitCloner) validateGitInstalled() error {
	cmd := NewRemoteCommand(g.ssh, "which git")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git is not installed on the remote system")
	}
	return nil
}

func (g *GitCloner) ensureDirectory(path string) error {
	cmd := NewRemoteCommand(g.ssh, fmt.Sprintf("mkdir -p %s", path))
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func (g *GitCloner) IsGitRepository(path string) bool {
	cmd := NewRemoteCommand(g.ssh, fmt.Sprintf("test -d %s/.git", path))
	return cmd.Run() == nil
}

func (g *GitCloner) GetCurrentCommit(repoPath string) (string, error) {
	cmd := NewRemoteCommand(g.ssh, fmt.Sprintf(
		"cd %s && git rev-parse HEAD",
		repoPath,
	))
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (g *GitCloner) validateOptions(opts CloneOptions) error {
	if opts.URL == "" {
		return fmt.Errorf("URL is required")
	}

	// Basic git URL format validation
	if !strings.HasPrefix(opts.URL, "git@") && !strings.HasPrefix(opts.URL, "https://") {
		return fmt.Errorf("invalid git URL format: must start with git@ or https://")
	}

	// Branch name validation
	if opts.Branch != "" && strings.ContainsAny(opts.Branch, " \t\n") {
		return fmt.Errorf("invalid branch name: must not contain whitespace")
	}

	return nil
}

func (g *GitCloner) buildCloneCommand(opts CloneOptions) string {
	var cmdParts []string

	cmdParts = append(cmdParts, "git -c advice.detachedHead=false clone --porcelain")

	if opts.Branch != "" {
		cmdParts = append(cmdParts, "-b", opts.Branch)
	}

	if opts.Depth > 0 {
		cmdParts = append(cmdParts, fmt.Sprintf("--depth=%d", opts.Depth))
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
