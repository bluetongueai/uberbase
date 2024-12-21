package pkg

import (
	"golang.org/x/crypto/ssh"
	"strings"
	"sync"
)

var _ SSHClientInterface = &MockSSHClient{}

type MockSSHClient struct {
	mu            sync.RWMutex
	commands      []string
	outputs       map[string]string
	errors        map[string]error
	customHandlers map[string]func(string) (string, error)
}

func NewMockSSHClient() *MockSSHClient {
	return &MockSSHClient{
		commands: make([]string, 0),
		outputs:  make(map[string]string),
		errors:   make(map[string]error),
		customHandlers: make(map[string]func(string) (string, error)),
	}
}

func (m *MockSSHClient) SetOutput(cmd, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.outputs[cmd] = output
}

func (m *MockSSHClient) SetError(cmd string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errors[cmd] = err
}

func (m *MockSSHClient) GetCommands() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return append([]string{}, m.commands...)
}

func (m *MockSSHClient) Connect() error {
	return nil
}

func (m *MockSSHClient) Close() error {
	return nil
}

func (m *MockSSHClient) GetClient() *ssh.Client {
	return nil
}

func (m *MockSSHClient) RunCommand(cmd string) (string, error) {
	m.mu.Lock()
	m.commands = append(m.commands, cmd)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// Check for custom handlers first
	for pattern, handler := range m.customHandlers {
		if strings.Contains(cmd, pattern) {
			return handler(cmd)
		}
	}
	
	// Check for exact matches
	if err, ok := m.errors[cmd]; ok {
		return "", err
	}
	if out, ok := m.outputs[cmd]; ok {
		return out, nil
	}

	// Check for partial matches
	for pattern, err := range m.errors {
		if strings.Contains(cmd, pattern) {
			return "", err
		}
	}
	for pattern, out := range m.outputs {
		if strings.Contains(cmd, pattern) {
			return out, nil
		}
	}

	return "", nil
}

func (m *MockSSHClient) WriteFile(path string, data []byte, perm uint32) error {
	m.mu.Lock()
	m.commands = append(m.commands, "write "+path)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()
	if err, ok := m.errors["write "+path]; ok {
		return err
	}
	return nil
}

func (m *MockSSHClient) ReadFile(path string) ([]byte, error) {
	m.mu.Lock()
	m.commands = append(m.commands, "read "+path)
	m.mu.Unlock()

	m.mu.RLock()
	defer m.mu.RUnlock()
	if err, ok := m.errors["read "+path]; ok {
		return nil, err
	}
	if out, ok := m.outputs["read "+path]; ok {
		return []byte(out), nil
	}
	return []byte{}, nil
}

func (m *MockSSHClient) SetCustomHandler(cmd string, handler func(string) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.customHandlers[cmd] = handler
}
