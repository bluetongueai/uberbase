package ssh

import (
	"fmt"
	"os"

	"golang.org/x/crypto/ssh"
)

type SSHKeySource int

const (
	Environment SSHKeySource = iota
	File
)

type SSHKey struct {
	Source  SSHKeySource
	loaded  bool
	envKey  string
	fileKey string
	auth    ssh.AuthMethod
}

func (k *SSHKey) Load() (ssh.AuthMethod, error) {
	if k.loaded {
		return k.auth, nil
	}

	if k.Source == Environment {
		return k.loadFromEnvironment()
	}
	if k.Source == File {
		return k.loadFromFile()
	}

	return nil, fmt.Errorf("invalid key source: %d", k.Source)
}

func (k *SSHKey) IsLoaded() bool {
	return k.loaded
}

func (k *SSHKey) loadFromFile() (ssh.AuthMethod, error) {
	keyBytes, err := os.ReadFile(k.fileKey)
	if err != nil {
		return nil, err
	}
	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return nil, err
	}
	k.loaded = true
	return ssh.PublicKeys(signer), nil
}

func (k *SSHKey) loadFromEnvironment() (ssh.AuthMethod, error) {
	keyData := os.Getenv(k.envKey)
	if keyData == "" {
		return nil, fmt.Errorf("environment variable %s is not set", k.envKey)
	}
	signer, err := ssh.ParsePrivateKey([]byte(keyData))
	if err != nil {
		return nil, err
	}
	k.loaded = true
	return ssh.PublicKeys(signer), nil
}
