package deploy

import (
	"fmt"
	"os"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
)

type GitManager struct {
	localExecutor  *core.LocalExecutor
	remoteExecutor *core.RemoteExecutor
	localWorkDir   string
	remoteWorkDir  string
	remote         string
	sha            string
}

func NewGitManager(localExecutor *core.LocalExecutor, remoteExecutor *core.RemoteExecutor, localWorkDir, remoteWorkDir string) (*GitManager, error) {
	logging.Logger.Debug("Creating new GitManager")
	manager := &GitManager{
		localExecutor:  localExecutor,
		remoteExecutor: remoteExecutor,
		localWorkDir:   localWorkDir,
		remoteWorkDir:  remoteWorkDir,
	}
	remote, err := manager.GetCurrentRepoURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get current repo URL: %w", err)
	}
	sha, err := manager.GetCurrentCommit()
	if err != nil {
		return nil, fmt.Errorf("failed to get current commit: %w", err)
	}
	manager.remote = remote
	manager.sha = sha
	return manager, nil
}

func (g *GitManager) GetCurrentRepoURL() (string, error) {
	remote, err := g.localExecutor.Exec(fmt.Sprintf("cd %s && git config --get remote.origin.url", g.localWorkDir))
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}
	return strings.TrimSpace(string(remote)), nil
}

func (g *GitManager) GetCurrentCommit() (string, error) {
	sha, err := g.localExecutor.Exec(fmt.Sprintf("cd %s && git rev-parse HEAD", g.localWorkDir))
	if err != nil {
		return "", fmt.Errorf("failed to get current commit: %w", err)
	}
	return strings.TrimSpace(string(sha)), nil
}

func (g *GitManager) Fetch() error {
	var cmd string

	if _, err := os.Stat(g.remoteWorkDir); os.IsNotExist(err) {
		logging.Logger.Debugf("Cloning repository: %s", g.remote)
		cmd = fmt.Sprintf("git clone %s %s", g.remote, g.remoteWorkDir)
	} else {
		logging.Logger.Debugf("Updating repository: %s", g.remote)
		cmd = fmt.Sprintf("git -C %s pull", g.remoteWorkDir)
	}

	_, err := g.remoteExecutor.Exec(cmd)
	if err != nil {
		logging.Logger.Errorf("Failed to fetch repository: %v", err)
		return fmt.Errorf("failed to fetch repository: %w", err)
	}

	logging.Logger.Debugf("Repository cloned successfully: %s", g.remote)
	return nil
}
