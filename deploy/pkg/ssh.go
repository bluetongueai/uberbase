package pkg

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	Host     string
	User     string
	Port     int
	client   *ssh.Client
	keyFile  string
	keyData  string
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

func (s *SSHClient) Connect() error {
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			s.publicKeyAuth(),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:        10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", s.Host, s.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
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

	// Try key data from environment first
	if s.keyData != "" {
		key = []byte(s.keyData)
	} else if s.keyFile != "" {
		// Then try key file
		key, err = os.ReadFile(s.keyFile)
		if err != nil {
			return nil
		}
	} else {
		// Finally try default key locations
		for _, defaultKey := range []string{"~/.ssh/id_rsa", "~/.ssh/id_ed25519"} {
			expanded := os.ExpandEnv(defaultKey)
			key, err = os.ReadFile(expanded)
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
