package pkg

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

func TestRemoteCommand(t *testing.T) {
	t.Run("Basic Command", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo hello")

		mock.SetOutput("echo hello", "hello")

		if err := cmd.Run(); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		commands := mock.GetCommands()
		if commands[0] != "echo hello" {
			t.Errorf("Expected command %q, got %q", "echo hello", commands[0])
		}
	})

	t.Run("Command Output", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "ls")

		expectedOutput := "file1\nfile2\nfile3"
		mock.SetOutput("ls", expectedOutput)

		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Output failed: %v", err)
		}

		if string(output) != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, string(output))
		}
	})

	t.Run("Command Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "invalid-command")

		mock.SetError("invalid-command", fmt.Errorf("command not found"))

		if err := cmd.Run(); err == nil {
			t.Error("Expected error for invalid command")
		}
	})

	t.Run("Custom Output Writer", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo hello")

		var stdout bytes.Buffer
		cmd.SetStdout(&stdout)

		mock.SetOutput("echo hello", "hello")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Output failed: %v", err)
		}

		if string(output) != "hello" {
			t.Errorf("Expected output %q, got %q", "hello", string(output))
		}
	})

	t.Run("Combined Output", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo hello && echo error >&2")

		mock.SetOutput("echo hello && echo error >&2", "hello\nerror")
		
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("CombinedOutput failed: %v", err)
		}

		expectedOutput := "hello\nerror"
		if string(output) != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, string(output))
		}
	})

	t.Run("Custom Error Writer", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo error >&2")

		var stderr bytes.Buffer
		cmd.SetStderr(&stderr)

		mock.SetOutput("echo error >&2", "error")
		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Output failed: %v", err)
		}

		if string(output) != "error" {
			t.Errorf("Expected output %q, got %q", "error", string(output))
		}
	})

	t.Run("Output After Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "failing-command")

		mock.SetError("failing-command", fmt.Errorf("command failed"))

		_, err := cmd.Output()
		if err == nil {
			t.Error("Expected error from Output() after command fails")
		}
	})
}

func TestRemoteCommand_InputOutput(t *testing.T) {
	t.Run("Large Output Handling", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "cat largefile")

		// Generate large output
		largeOutput := strings.Repeat("x", 1024*1024) // 1MB of data
		mock.SetOutput("cat largefile", largeOutput)

		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to handle large output: %v", err)
		}
		if len(output) != len(largeOutput) {
			t.Errorf("Expected output length %d, got %d", len(largeOutput), len(output))
		}
	})

	t.Run("Multiple Writers", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo test")

		var stdout1, stdout2 bytes.Buffer
		multiWriter := io.MultiWriter(&stdout1, &stdout2)
		cmd.SetStdout(multiWriter)

		mock.SetOutput("echo test", "test output")
		if err := cmd.Run(); err != nil {
			t.Errorf("Run failed: %v", err)
		}

		if stdout1.String() != stdout2.String() {
			t.Error("Writers received different output")
		}
	})

	t.Run("Nil Writer Handling", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo test")

		cmd.SetStdout(nil)
		cmd.SetStderr(nil)

		if err := cmd.Run(); err != nil {
			t.Errorf("Failed to handle nil writers: %v", err)
		}
	})
}

func TestRemoteCommand_ErrorHandling(t *testing.T) {
	t.Run("Command Timeout", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "sleep 10")

		mock.SetError("sleep 10", fmt.Errorf("command timed out"))

		if err := cmd.Run(); err == nil {
			t.Error("Expected timeout error")
		}
	})

	t.Run("Invalid Command Format", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "")

		err := cmd.Run()
		if err == nil {
			t.Error("Expected error for empty command")
		}
		if !strings.Contains(err.Error(), "command cannot be empty") {
			t.Errorf("Expected 'command cannot be empty' error, got: %v", err)
		}
	})

	t.Run("SSH Connection Lost", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo test")

		mock.SetError("echo test", fmt.Errorf("connection reset"))

		if err := cmd.Run(); err == nil {
			t.Error("Expected connection error")
		}
	})
}

func TestRemoteCommand_OutputFormatting(t *testing.T) {
	t.Run("Handle Special Characters", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo 'special chars: $@#%'")

		expectedOutput := "special chars: $@#%"
		mock.SetOutput("echo 'special chars: $@#%'", expectedOutput)

		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to handle special characters: %v", err)
		}
		if string(output) != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, string(output))
		}
	})

	t.Run("Handle Multiline Output", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "ls -l")

		expectedOutput := "file1\nfile2\nfile3\n"
		mock.SetOutput("ls -l", expectedOutput)

		output, err := cmd.Output()
		if err != nil {
			t.Errorf("Failed to handle multiline output: %v", err)
		}
		if string(output) != expectedOutput {
			t.Errorf("Expected output %q, got %q", expectedOutput, string(output))
		}
	})
}

func TestRemoteCommand_Validation(t *testing.T) {
	t.Run("Command Length Limit", func(t *testing.T) {
		mock := NewMockSSHClient()
		longCommand := strings.Repeat("x", 10000)
		cmd := NewRemoteCommand(mock, longCommand)

		if err := cmd.Run(); err == nil {
			t.Error("Expected error for extremely long command")
		}
	})

	t.Run("Invalid Characters in Command", func(t *testing.T) {
		mock := NewMockSSHClient()
		cmd := NewRemoteCommand(mock, "echo \x00test") // Null byte

		if err := cmd.Run(); err == nil {
			t.Error("Expected error for invalid characters")
		}
	})
}
