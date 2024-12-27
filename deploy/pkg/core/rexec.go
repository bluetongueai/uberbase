package core

import (
	"fmt"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host    string
	User    string
	Port    int
	KeyFile string
	KeyData string
}

// SSHConnection represents a persistent SSH connection
type SSHConnection struct {
	client *ssh.Client
	addr   string
	user   string
	auth   ssh.AuthMethod
	mu     sync.Mutex
	closed bool
}

// NewSession creates a new persistent SSH connection
func NewSession(config SSHConfig) (*SSHConnection, error) {
	var auth ssh.AuthMethod
	var err error
	if config.KeyData == "" {
		auth, err = ParsePrivateKeyFile(config.KeyFile)
		if err != nil {
			return nil, err
		}
	} else {
		auth, err = ParsePrivateKey(config.KeyData)
		if err != nil {
			return nil, err
		}
	}
	conn := &SSHConnection{
		addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
		user: config.User,
		auth: auth,
	}

	if err := conn.connect(); err != nil {
		return nil, err
	}

	return conn, nil
}

// connect establishes the SSH connection
func (c *SSHConnection) connect() error {
	cfg := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			c.auth,
		},
		HostKeyCallback: ssh.HostKeyCallback(
			func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		),
	}

	client, err := ssh.Dial("tcp", c.addr, cfg)
	if err != nil {
		return err
	}

	c.client = client
	return nil
}

// Exec executes a command over the persistent connection
func (c *SSHConnection) Exec(cmd string) ([]byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil, fmt.Errorf("connection is closed")
	}

	session, err := c.client.NewSession()
	if err != nil {
		// Try to reconnect once if the session creation fails
		if err := c.connect(); err != nil {
			return nil, err
		}
		session, err = c.client.NewSession()
		if err != nil {
			return nil, err
		}
	}
	defer session.Close()

	return session.CombinedOutput(cmd)
}

// Close closes the SSH connection
func (c *SSHConnection) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}

	c.closed = true
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// ParsePrivateKey parses an SSH private key string and returns an ssh.AuthMethod
func ParsePrivateKey(privateKey string) (ssh.AuthMethod, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return nil, err
	}
	return ssh.PublicKeys(signer), nil
}

// ParsePrivateKeyFile reads and parses an SSH private key file and returns an ssh.AuthMethod
func ParsePrivateKeyFile(keyPath string) (ssh.AuthMethod, error) {
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return ParsePrivateKey(string(keyBytes))
}

// Add this method to the SSHConnection struct
func (c *SSHConnection) Host() string {
	return c.addr
}
