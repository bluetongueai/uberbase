package pkg

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type RemoteCommand struct {
	client  *SSHClient
	command string
	stdout  io.Writer
	stderr  io.Writer
}

func NewRemoteCommand(client *SSHClient, command string) *RemoteCommand {
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

func (r *RemoteCommand) Run() error {
	session, err := r.client.GetClient().NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	session.Stdout = r.stdout
	session.Stderr = r.stderr

	return session.Run(r.command)
}

func (r *RemoteCommand) Output() ([]byte, error) {
	var stdout bytes.Buffer
	r.SetStdout(&stdout)

	err := r.Run()
	return stdout.Bytes(), err
}

func (r *RemoteCommand) CombinedOutput() ([]byte, error) {
	var combined bytes.Buffer
	r.SetStdout(&combined)
	r.SetStderr(&combined)

	err := r.Run()
	return combined.Bytes(), err
}
