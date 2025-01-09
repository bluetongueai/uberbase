package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host string
	User string
	Port int
	Key  SSHKey
}

type SSHSession struct {
	client *ssh.Client
	addr   string
	user   string
	auth   ssh.AuthMethod
	closed bool
}

func NewSession(config SSHConfig) (*SSHSession, error) {
	auth, err := config.Key.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load SSH key: %w", err)
	}

	logging.Logger.Debug("Creating SSH session")
	conn := &SSHSession{
		addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
		user: config.User,
		auth: auth,
	}
	return conn, nil
}

func (c *SSHSession) Connect() (*ssh.Session, error) {
	if c.client == nil {
		cfg := &ssh.ClientConfig{
			User: c.user,
			Auth: []ssh.AuthMethod{
				c.auth,
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}

		client, err := ssh.Dial("tcp", c.addr, cfg)
		if err != nil {
			c.closed = true
			logging.Logger.Errorf("SSH connection failed: %v", err)
			return nil, err
		}
		c.client = client
	}

	session, err := c.client.NewSession()
	if err != nil {
		c.closed = true
		c.client = nil
		return nil, err
	}

	if c.closed {
		c.closed = false
	}
	return session, nil
}

func (c *SSHSession) ExecuteCommand(cmd string) (string, error) {
	session, err := c.Connect()
	if err != nil {
		return "", err
	}
	defer session.Close()

	logging.Logger.Infof("%s: \033[34m%s\033[0m", c.addr, cmd)

	// Create buffers for stdout and stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Run the command
	err = session.Run(cmd)

	// If there's stderr output, log it in red regardless of error
	if stderrStr := stderr.String(); stderrStr != "" {
		logging.Logger.Infof("\033[31m%s\033[0m", stderrStr)
	}

	if err != nil {
		c.closed = true
		c.client = nil
		return "", err
	}

	return stdout.String(), nil
}

func (c *SSHSession) Close() {
	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	c.closed = true
}

func (c *SSHSession) IsClosed() bool {
	return c.closed
}

func (s *SSHSession) TransferFile(localPath, remotePath string) error {
	// Create new SFTP client
	sftpClient, err := sftp.NewClient(s.client)
	if err != nil {
		return fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftpClient.Close()

	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Ensure the remote directory exists
	remoteDir := filepath.Dir(remotePath)
	err = sftpClient.MkdirAll(remoteDir)
	if err != nil {
		return fmt.Errorf("failed to create remote directory: %w", err)
	}

	// Create remote file
	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Copy file contents
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
