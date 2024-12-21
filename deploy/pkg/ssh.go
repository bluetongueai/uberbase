package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHSession interface {
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}

type SSHClient struct {
	Host     string
	User     string
	Port     int
	client   *ssh.Client
	keyFile  string
	keyData  string
	timeout  time.Duration
	newSession func() (SSHSession, error) // For testing
}

type SSHClientInterface interface {
	Connect() error
	Close() error
	GetClient() *ssh.Client
	RunCommand(cmd string) (string, error)
	WriteFile(path string, data []byte, perm uint32) error
	ReadFile(path string) ([]byte, error)
}

func NewSSHClient(host string, user string, port int, keyFile string, keyData string) *SSHClient {
	return &SSHClient{
		Host:     host,
			User:     user,
			Port:     port,
			keyFile:  keyFile,
			keyData:  keyData,
	}
}

func (s *SSHClient) validate() error {
	if s.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	if s.Port < 0 || s.Port > 65535 {
		return fmt.Errorf("invalid port: %d", s.Port)
	}
	if s.keyFile == "" && s.keyData == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil || !s.hasDefaultKeys(homeDir) {
			return fmt.Errorf("no authentication method available")
		}
	}
	return nil
}

func (s *SSHClient) hasDefaultKeys(homeDir string) bool {
	for _, key := range []string{"id_rsa", "id_ed25519"} {
		keyPath := filepath.Join(homeDir, ".ssh", key)
		if _, err := os.Stat(keyPath); err == nil {
			return true
		}
	}
	return false
}

func (s *SSHClient) Connect() error {
	if s.client != nil {
		return fmt.Errorf("already connected")
	}

	if err := s.validate(); err != nil {
		return err
	}

	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			s.publicKeyAuth(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:        s.timeout,
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	s.client = client
	return nil
}

func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSHClient) publicKeyAuth() ssh.AuthMethod {
	var key []byte
	var err error

	// Try key file first and ONLY if specified
	if s.keyFile != "" {
		key, err = os.ReadFile(s.keyFile)
		if err != nil {
			return nil  // Return nil if specified file doesn't exist
		}
	} else if s.keyData != "" {
		// Try key data only if no keyFile specified
		key = []byte(s.keyData)
	} else {
		// Try default key locations only if neither keyFile nor keyData specified
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		for _, defaultKey := range []string{"id_rsa", "id_ed25519"} {
			keyPath := filepath.Join(homeDir, ".ssh", defaultKey)
			key, err = os.ReadFile(keyPath)
			if err == nil {
				break
			}
		}
	}

	if key == nil {
		return nil
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		return nil
	}

	return ssh.PublicKeys(signer)
}

func (s *SSHClient) GetClient() *ssh.Client {
	return s.client
}

func (s *SSHClient) RunCommand(cmd string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("not connected")
	}
	
	session, err := s.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run command: %w", err)
	}

	return string(output), nil
}
