package ssh

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
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

func NewSSHKey(source SSHKeySource, envKey string, fileKey string) *SSHKey {
	return &SSHKey{
		Source:  source,
		envKey:  envKey,
		fileKey: fileKey,
	}
}

func (k *SSHKey) Load() (ssh.AuthMethod, error) {
	if k.loaded {
		return k.auth, nil
	}

	var err error
	if k.Source == Environment {
		k.auth, err = k.loadFromEnvironment()
	} else if k.Source == File {
		k.auth, err = k.loadFromFile()
	} else {
		return nil, fmt.Errorf("invalid key source: %d", k.Source)
	}

	if err != nil {
		return nil, err
	}

	k.loaded = true
	return k.auth, nil
}

func (k *SSHKey) IsLoaded() bool {
	return k.loaded
}

func (k *SSHKey) loadFromFile() (ssh.AuthMethod, error) {
	// turn path into absolute path while resolving ~ and etc
	relPath := k.fileKey
	if strings.HasPrefix(relPath, "~/") {
		relPath = strings.Replace(relPath, "~", os.Getenv("HOME"), 1)
	}
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return nil, err
	}

	logging.Logger.Debug("Loading SSH key from file", "path", absPath)

	keyBytes, err := os.ReadFile(absPath)
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
