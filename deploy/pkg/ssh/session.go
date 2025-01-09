package ssh

import (
	"fmt"

	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	"golang.org/x/crypto/ssh"
)

type SSHConfig struct {
	Host string
	User string
	Port int
	Key  SSHKey
}

type SSHSession struct {
	session *ssh.Session
	addr    string
	user    string
	auth    ssh.AuthMethod
	closed  bool
}

func NewSession(config SSHConfig) (*SSHSession, error) {
	if !config.Key.IsLoaded() {
		logging.Logger.Debug("Loading key")
		if _, err := config.Key.Load(); err != nil {
			logging.Logger.Debugf("Failed to load key: %v", err)
			return nil, err
		}
	} else {
		logging.Logger.Debug("Key loaded")
	}

	logging.Logger.Debug("Creating closed SSH session")
	conn := &SSHSession{
		addr: fmt.Sprintf("%s:%d", config.Host, config.Port),
		user: config.User,
		auth: config.Key.auth,
	}
	return conn, nil
}

func (c *SSHSession) Connect() (*ssh.Session, error) {
	logging.Logger.Info("Connecting SSH session")

	cfg := &ssh.ClientConfig{
		User: c.user,
		Auth: []ssh.AuthMethod{
			c.auth,
		},
	}

	client, err := ssh.Dial("tcp", c.addr, cfg)
	if err != nil {
		logging.Logger.Warnf("Failed to connect to SSH: %v", err)
		return nil, err
	}

	session, err := client.NewSession()
	if err != nil {
		logging.Logger.Warnf("Failed to create session: %v", err)
		return nil, err
	}

	c.session = session
	c.closed = false
	logging.Logger.Info("Connected to SSH")
	return session, nil
}

func (c *SSHSession) Close() {
	logging.Logger.Info("Closing SSH connection")
	c.session.Close()
	c.closed = true
}

func (c *SSHSession) IsClosed() bool {
	return c.closed
}
