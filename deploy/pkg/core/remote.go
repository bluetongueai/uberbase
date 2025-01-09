package core

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	bt_ssh "github.com/bluetongueai/uberbase/deploy/pkg/ssh"
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
	_, err := p.session.Connect()
	if err != nil {
		logging.Logger.Debugf("Failed to connect to remote server: %v", err)
		return false
	}
	logging.Logger.Debug("Connected to remote server")
	return true
}

func (p *RemoteExecutor) Verify() error {
	if !p.installer.HasGit() {
		logging.Logger.Debug("Git is not installed")
		if err := p.installer.InstallGit(); err != nil {
			return fmt.Errorf("failed to install git: %w", err)
		}
	}
	if !p.installer.HasDocker() {
		if !p.installer.HasPodman() {
			logging.Logger.Debug("Neither docker nor podman is installed")
			if err := p.installer.InstallContainerRuntime(); err != nil {
				return fmt.Errorf("failed to install container runtime: %w", err)
			}
		} else {
			if !p.installer.HasPodmanCompose() {
				logging.Logger.Debug("Podman compose is not installed")
				if err := p.installer.InstallContainerRuntime(); err != nil {
					return fmt.Errorf("failed to install podman compose: %w", err)
				}
			}
		}
	}
	return nil
}

func (p *RemoteExecutor) Exec(cmd string) (string, error) {
	if p.session.IsClosed() {
		_, err := p.session.Connect()
		if err != nil {
			return "", fmt.Errorf("failed to connect to remote server: %w", err)
		}
	}

	for attempt := 1; attempt <= p.maxRetries; attempt++ {
		output, err := p.execWithoutRetry(cmd)
		if err == nil {
			return string(output), nil
		}

		// Only retry on lock conflicts or transport errors
		if (strings.Contains(err.Error(), "could not get lock") ||
			isTransportError(err)) && attempt < p.maxRetries {
			logging.Logger.Debugf("Retrying command due to transport error (attempt %d): %v", attempt, err)
			time.Sleep(p.retryDelay)
			continue
		}

		return "", fmt.Errorf("failed to execute %s: %w", cmd, err)
	}
	return "", nil
}

// isTransportError checks if the error is related to transport/connection issues
func isTransportError(err error) bool {
	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "no route to host") ||
		strings.Contains(errStr, "connection reset by peer") ||
		strings.Contains(errStr, "broken pipe") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "transport") ||
		strings.Contains(errStr, "connection closed")
}

func (c *RemoteExecutor) execWithoutRetry(cmd string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Execute command on existing session
	output, err := c.session.ExecuteCommand(cmd)
	if err != nil && isTransportError(err) {
		// Only try to reconnect if it's a transport error
		if _, reconnErr := c.session.Connect(); reconnErr == nil {
			output, err = c.session.ExecuteCommand(cmd)
		}
	}
	return output, err
}

// SendFile transfers a local file to the remote server
func (p *RemoteExecutor) SendFile(localPath, remotePath string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	err := p.session.TransferFile(localPath, remotePath)
	if err != nil && isTransportError(err) {
		// Only try to reconnect if it's a transport error
		if _, reconnErr := p.session.Connect(); reconnErr == nil {
			err = p.session.TransferFile(localPath, remotePath)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to transfer file from %s to %s: %w", localPath, remotePath, err)
	}

	return nil
}
