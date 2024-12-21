package pkg

import (
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	maxCommandLength = 8192 // Maximum length for command string
)

type RemoteCommand struct {
	client  SSHClientInterface
	command string
	stdout  io.Writer
	stderr  io.Writer
}

func NewRemoteCommand(client SSHClientInterface, command string) *RemoteCommand {
	return &RemoteCommand{
		client:  client,
		command: command,
		stdout:  os.Stdout,
		stderr:  os.Stderr,
	}
}

func (r *RemoteCommand) SetStdout(w io.Writer) {
	r.stdout = w
}

func (r *RemoteCommand) SetStderr(w io.Writer) {
	r.stderr = w
}

func (r *RemoteCommand) validateCommand() error {
	if r == nil {
		return fmt.Errorf("remote command is nil")
	}
	if r.command == "" {
		return fmt.Errorf("command cannot be empty")
	}
	if len(r.command) > maxCommandLength {
		return fmt.Errorf("command exceeds maximum length of %d characters", maxCommandLength)
	}
	// Check for null bytes or other invalid characters
	if strings.Contains(r.command, "\x00") {
		return fmt.Errorf("command contains invalid characters")
	}
	return nil
}

func (r *RemoteCommand) Run() error {
	if err := r.validateCommand(); err != nil {
		return err
	}
	_, err := r.client.RunCommand(r.command)
	if err != nil {
		return fmt.Errorf("failed to run command: %w", err)
	}
	return nil
}

func (r *RemoteCommand) Output() ([]byte, error) {
	if err := r.validateCommand(); err != nil {
		return nil, err
	}
	output, err := r.client.RunCommand(r.command)
	if err != nil {
		return nil, err
	}
	return []byte(output), nil
}

func (r *RemoteCommand) CombinedOutput() ([]byte, error) {
	if err := r.validateCommand(); err != nil {
		return nil, err
	}
	output, err := r.client.RunCommand(r.command)
	if err != nil {
		return nil, err
	}
	return []byte(output), nil
}
