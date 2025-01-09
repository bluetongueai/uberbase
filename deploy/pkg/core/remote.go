package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	bt_ssh "github.com/bluetongueai/uberbase/deploy/pkg/ssh"
	"golang.org/x/crypto/ssh"
)

type RemoteExecutor struct {
	installer  *Installer
	session    *bt_ssh.SSHSession
	mu         sync.Mutex
	maxRetries int
	retryDelay time.Duration
}

func NewRemoteExecutor(config bt_ssh.SSHConfig) (*RemoteExecutor, error) {
	session, err := bt_ssh.NewSession(config)
	if err != nil {
		return nil, err
	}
	executor := &RemoteExecutor{
		session:    session,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}
	executor.installer = NewInstaller(executor)
	return executor, nil
}

func (p *RemoteExecutor) Test() bool {
	logging.Logger.Debug("Testing remote connection")
	session, err := p.session.Connect()
	if err != nil {
		logging.Logger.Debugf("Failed to connect to remote server: %v", err)
		return false
	}
	defer session.Close()
	logging.Logger.Debug("Connected to remote server")
	session.Close()
	return true
}

func (p *RemoteExecutor) Verify() error {
	if !p.installer.HasGit() {
		logging.Logger.Debug("Git is not installed")
		logging.Logger.Info("Installing git")
		if err := p.installer.InstallGit(); err != nil {
			return fmt.Errorf("failed to install git: %w", err)
		}
		logging.Logger.Info("Git installed")
	}
	if !p.installer.HasDocker() {
		if !p.installer.HasPodman() {
			logging.Logger.Debug("Neither docker nor podman is installed")
			logging.Logger.Info("Installing container runtime")
			if err := p.installer.InstallContainerRuntime(); err != nil {
				return fmt.Errorf("failed to install container runtime: %w", err)
			}
			logging.Logger.Info("Container runtime installed")
		} else {
			if !p.installer.HasPodmanCompose() {
				logging.Logger.Debug("Podman compose is not installed")
				logging.Logger.Info("Installing podman compose")
				if err := p.installer.InstallContainerRuntime(); err != nil {
					return fmt.Errorf("failed to install podman compose: %w", err)
				}
				logging.Logger.Info("Podman compose installed")
			}
		}
	}
	return nil
}

func (p *RemoteExecutor) Exec(cmd string) (string, error) {
	if p.session.IsClosed() {
		logging.Logger.Debug("Session is closed, connecting")
		session, err := p.session.Connect()
		if err != nil {
			logging.Logger.Debugf("Failed to connect to remote server: %v", err)
			return "", err
		}
		defer session.Close()
	}

	for attempt := 1; attempt <= p.maxRetries; attempt++ {
		logging.Logger.Infof("Executing command (attempt %d/%d): %s", attempt, p.maxRetries, cmd)
		output, err := p.execWithoutRetry(cmd)
		if err == nil {
			logging.Logger.Infof("%s completed successfully", cmd)
			return string(output), nil
		}

		if strings.Contains(err.Error(), "could not get lock") && attempt < p.maxRetries {
			logging.Logger.Warnf("Could not get lock, retrying %s... (%d/%d)", cmd, attempt, p.maxRetries)
			time.Sleep(p.retryDelay)
			continue
		}

		logging.Logger.Errorf("Failed to execute %s: %v", cmd, err)
		return "", fmt.Errorf("failed to execute %s: %w", cmd, err)
	}
	return "", nil
}

// Exec executes a command over the persistent connection
func (c *RemoteExecutor) execWithoutRetry(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var session *ssh.Session
	if c.session.IsClosed() {
		logging.Logger.Info("Connection is closed, reconnecting")
		var err error
		session, err = c.session.Connect()
		if err != nil {
			logging.Logger.Warnf("Failed to reconnect: %v", err)
			return "", err
		}
	}

	defer c.session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
