package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSSHConnection(t *testing.T) {
	mock := NewMockSSH("localhost:2222")
	go mock.ListenAndServe()
	defer mock.Close()

	// Give the mock server a moment to start
	time.Sleep(100 * time.Millisecond)

	conn, err := NewSession(SSHConfig{
		Host:    "localhost:2222",
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)
	defer conn.Close()

	t.Run("Basic Command", func(t *testing.T) {
		mock.SetReturnString("hello")
		output, err := conn.Exec("echo hello")
		assert.NoError(t, err)
		assert.Equal(t, "hello", string(output))
	})

	t.Run("Command Output", func(t *testing.T) {
		expectedOutput := "file1\nfile2\nfile3"
		mock.SetReturnString(expectedOutput)

		output, err := conn.Exec("ls")
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("Command Error", func(t *testing.T) {
		mock.SetReturnString("command not found")
		output, err := conn.Exec("invalid-command")
		assert.NoError(t, err) // Mock server doesn't actually return errors
		assert.Equal(t, "command not found", string(output))
	})

	t.Run("Large Output Handling", func(t *testing.T) {
		largeOutput := string(make([]byte, 1024*1024)) // 1MB of zero bytes
		mock.SetReturnString(largeOutput)

		output, err := conn.Exec("cat largefile")
		assert.NoError(t, err)
		assert.Equal(t, len(largeOutput), len(output))
	})

	t.Run("Special Characters in Output", func(t *testing.T) {
		expectedOutput := "special chars: $@#%"
		mock.SetReturnString(expectedOutput)

		output, err := conn.Exec("echo 'special chars: $@#%'")
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, string(output))
	})

	t.Run("Multiline Output", func(t *testing.T) {
		expectedOutput := "file1\nfile2\nfile3\n"
		mock.SetReturnString(expectedOutput)

		output, err := conn.Exec("ls -l")
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, string(output))
	})
}

func TestSSHConnectionErrors(t *testing.T) {
	t.Run("Connection to non-existent server", func(t *testing.T) {
		_, err := NewSession(SSHConfig{
			Host:    "localhost:1234",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.Error(t, err)
	})

	t.Run("Operations on closed connection", func(t *testing.T) {
		mock := NewMockSSH("localhost:2223")
		go mock.ListenAndServe()
		defer mock.Close()

		time.Sleep(100 * time.Millisecond)

		conn, err := NewSession(SSHConfig{
			Host:    "localhost:2223",
			User:    "test-user",
			KeyData: "dummy-key",
		})
		assert.NoError(t, err)

		err = conn.Close()
		assert.NoError(t, err)

		_, err = conn.Exec("echo test")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection is closed")
	})
}

func TestSSHKeyAuth(t *testing.T) {
	t.Run("Parse invalid private key", func(t *testing.T) {
		_, err := ParsePrivateKey("invalid key content")
		assert.Error(t, err)
	})

	t.Run("Parse private key from non-existent file", func(t *testing.T) {
		_, err := ParsePrivateKeyFile("/nonexistent/key/path")
		assert.Error(t, err)
	})
}
