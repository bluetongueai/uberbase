package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupMockSSH(t *testing.T, port string) (*MockSSH, *SSHConnection) {
	mock := NewMockSSH("localhost:" + port)
	go mock.ListenAndServe()

	// Give the mock server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Create test connection
	conn, err := NewSession(SSHConfig{
		Host:    "localhost:" + port,
		User:    "test-user",
		KeyData: "dummy-key",
	})
	assert.NoError(t, err)

	return mock, conn
}

func TestPodmanInstaller(t *testing.T) {
	mock, conn := setupMockSSH(t, "2222")
	defer mock.Close()
	defer conn.Close()

	t.Run("Already Installed", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)
		mock.SetReturnString("/usr/bin/podman")
		mock.SetReturnString("/usr/bin/podman-compose")

		err := installer.EnsureInstalled()
		assert.NoError(t, err)
	})

	t.Run("Install on Ubuntu", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)

		// Mock responses for system requirement checks
		mock.SetReturnString("5.4.0")                                 // uname -r
		mock.SetReturnString("7723")                                  // free -m
		mock.SetReturnString("20000000")                              // df output
		mock.SetReturnString("NAME=\"Ubuntu\"\nVERSION_ID=\"20.04\"") // os-release
		mock.SetReturnString("")                                      // which podman (not found)
		mock.SetReturnString("")                                      // apt-get install success
		mock.SetReturnString("")                                      // which podman-compose (not found)
		mock.SetReturnString("")                                      // pip install success
		mock.SetReturnString("")                                      // podman-compose install success

		err := installer.EnsureInstalled()
		assert.NoError(t, err)
	})

	t.Run("Install on CentOS", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)

		mock.SetReturnString("5.4.0")                 // uname -r
		mock.SetReturnString("7723")                  // free -m
		mock.SetReturnString("20000000")              // df output
		mock.SetReturnString("NAME=\"CentOS Linux\"") // os-release
		mock.SetReturnString("")                      // which podman (not found)
		mock.SetReturnString("")                      // dnf install success
		mock.SetReturnString("")                      // which podman-compose (not found)
		mock.SetReturnString("")                      // pip install success
		mock.SetReturnString("")                      // podman-compose install success

		err := installer.EnsureInstalled()
		assert.NoError(t, err)
	})

	t.Run("Unsupported OS", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)
		mock.SetReturnString("5.4.0")               // uname -r
		mock.SetReturnString("7723")                // free -m
		mock.SetReturnString("20000000")            // df output
		mock.SetReturnString("NAME=\"Arch Linux\"") // os-release

		err := installer.EnsureInstalled()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported operating system")
	})
}

func TestPodmanInstaller_DetectOS(t *testing.T) {
	mock, conn := setupMockSSH(t, "2223")
	defer mock.Close()
	defer conn.Close()

	tests := []struct {
		name     string
		osOutput string
		want     string
		wantErr  bool
	}{
		{
			name:     "Ubuntu 20.04",
			osOutput: "NAME=\"Ubuntu\"\nVERSION_ID=\"20.04\"",
			want:     "ubuntu",
			wantErr:  false,
		},
		{
			name:     "Debian 11",
			osOutput: "NAME=\"Debian GNU/Linux\"\nVERSION_ID=\"11\"",
			want:     "debian",
			wantErr:  false,
		},
		{
			name:     "CentOS 8",
			osOutput: "NAME=\"CentOS Linux\"\nVERSION_ID=\"8\"",
			want:     "centos",
			wantErr:  false,
		},
		{
			name:     "Invalid OS",
			osOutput: "NAME=\"SomeOS\"\nVERSION_ID=\"1.0\"",
			want:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installer := NewPodmanInstaller(conn)
			mock.SetReturnString(tt.osOutput)

			got, err := installer.detectOS()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestPodmanInstaller_SystemRequirements(t *testing.T) {
	mock, conn := setupMockSSH(t, "2224")
	defer mock.Close()
	defer conn.Close()

	t.Run("Valid Requirements", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)
		mock.SetReturnString("5.4.0")    // uname -r
		mock.SetReturnString("7723")     // free -m
		mock.SetReturnString("20000000") // df output

		err := installer.checkSystemRequirements()
		assert.NoError(t, err)
	})

	t.Run("Insufficient Memory", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)
		mock.SetReturnString("5.4.0") // uname -r
		mock.SetReturnString("1024")  // free -m (insufficient)

		err := installer.checkSystemRequirements()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient memory")
	})

	t.Run("Old Kernel Version", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)
		mock.SetReturnString("2.6.32") // uname -r (too old)

		err := installer.checkSystemRequirements()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "kernel version")
	})

	t.Run("Insufficient Disk Space", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)
		mock.SetReturnString("5.4.0")   // uname -r
		mock.SetReturnString("7723")    // free -m
		mock.SetReturnString("5000000") // df output (insufficient)

		err := installer.checkSystemRequirements()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient disk space")
	})
}

func TestPodmanInstaller_InstallationRetries(t *testing.T) {
	mock, conn := setupMockSSH(t, "2225")
	defer mock.Close()
	defer conn.Close()

	t.Run("Retry on Package Manager Lock", func(t *testing.T) {
		installer := NewPodmanInstaller(conn)

		// System requirements checks
		mock.SetReturnString("5.4.0")                                 // uname -r
		mock.SetReturnString("7723")                                  // free -m
		mock.SetReturnString("20000000")                              // df output
		mock.SetReturnString("NAME=\"Ubuntu\"")                       // os-release
		mock.SetReturnString("")                                      // which podman (not found)
		mock.SetReturnString("could not get lock /var/lib/dpkg/lock") // First attempt fails
		mock.SetReturnString("")                                      // Second attempt succeeds
		mock.SetReturnString("")                                      // which podman-compose (not found)
		mock.SetReturnString("")                                      // pip install success
		mock.SetReturnString("")                                      // podman-compose install success

		err := installer.EnsureInstalled()
		assert.NoError(t, err)
	})
}
