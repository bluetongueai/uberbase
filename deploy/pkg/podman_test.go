package pkg

import (
	"fmt"
	"strings"
	"testing"
)

func TestPodmanInstaller(t *testing.T) {
	t.Run("Already Installed", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		mock.SetOutput("uname -r", "5.4.0")
		mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "7723")
		mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1", "20000000")
		
		mock.SetOutput("which podman", "/usr/bin/podman")
		mock.SetOutput("which podman-compose", "/usr/local/bin/podman-compose")

		if err := installer.EnsureInstalled(); err != nil {
			t.Errorf("EnsureInstalled failed: %v", err)
		}
	})

	t.Run("Install on Ubuntu", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		// Mock system requirements
		mock.SetOutput("uname -r", "5.4.0")
		mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "7723")
		mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1", "20000000")

		// Simulate podman not installed
		mock.SetError("which podman", fmt.Errorf("not found"))
		// Simulate Ubuntu
		mock.SetOutput("cat /etc/os-release", "NAME=\"Ubuntu\"\nVERSION_ID=\"20.04\"")
		// Simulate successful installation
		mock.SetOutput("sudo apt-get update && sudo apt-get install -y podman", "")
		// Simulate podman-compose not installed
		mock.SetError("which podman-compose", fmt.Errorf("not found"))
		// Simulate pip installation
		mock.SetOutput("which pip3 || ( sudo apt-get update && sudo apt-get install -y python3-pip )", "")
		mock.SetOutput("sudo pip3 install podman-compose", "")

		if err := installer.EnsureInstalled(); err != nil {
			t.Errorf("EnsureInstalled failed: %v", err)
		}
	})

	t.Run("Install on CentOS", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		mock.SetError("which podman", fmt.Errorf("not found"))
		mock.SetOutput("cat /etc/os-release", "NAME=\"CentOS Linux\"")
		mock.SetOutput("sudo dnf -y install 'dnf-command(copr)' && sudo dnf -y copr enable rhcontainerbot/container-selinux && sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/CentOS_8/devel:kubic:libcontainers:stable.repo && sudo dnf -y install podman", "")

		if err := installer.checkAndInstallPodman(); err != nil {
			t.Errorf("checkAndInstallPodman failed: %v", err)
		}
	})

	t.Run("Unsupported OS", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		mock.SetError("which podman", fmt.Errorf("not found"))
		mock.SetOutput("cat /etc/os-release", "NAME=\"Arch Linux\"")

		if err := installer.EnsureInstalled(); err == nil {
			t.Error("Expected error for unsupported OS")
		}
	})

	t.Run("OS Detection Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		mock.SetError("which podman", fmt.Errorf("not found"))
		mock.SetError("cat /etc/os-release", fmt.Errorf("file not found"))

		if err := installer.EnsureInstalled(); err == nil {
			t.Error("Expected error when OS detection fails")
		}
	})

	t.Run("Installation Command Error", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		mock.SetError("which podman", fmt.Errorf("not found"))
		mock.SetOutput("cat /etc/os-release", "NAME=\"Ubuntu\"")
		mock.SetError("sudo apt-get update && sudo apt-get install -y podman", fmt.Errorf("installation failed"))

		if err := installer.EnsureInstalled(); err == nil {
			t.Error("Expected error when installation command fails")
		}
	})
}

func TestPodmanInstaller_DetectOS(t *testing.T) {
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
			name:     "RHEL 8",
			osOutput: "NAME=\"Red Hat Enterprise Linux\"\nID=\"rhel\"\nVERSION_ID=\"8.4\"",
			want:     "rhel",
			wantErr:  false,
		},
		{
			name:     "Fedora 35",
			osOutput: "NAME=\"Fedora Linux\"\nVERSION_ID=\"35\"",
			want:     "fedora",
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
			mock := NewMockSSHClient()
			installer := NewPodmanInstaller(mock)

			// Set up system requirement mocks
			mock.SetOutput("uname -r", "5.4.0")
			mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "4096")
			mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail /", "20000000")
			mock.SetOutput("cat /etc/os-release", tt.osOutput)

			got, err := installer.detectOS()
			if (err != nil) != tt.wantErr {
				t.Errorf("detectOS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("detectOS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPodmanInstaller_InstallationRetries(t *testing.T) {
	t.Run("Retry on Package Manager Lock", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		// Mock system requirements
		mock.SetOutput("uname -r", "5.4.0")
		mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "7723")
		mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1", "20000000")

		// Mock OS detection
		mock.SetOutput("cat /etc/os-release", "NAME=\"Ubuntu\"")
		
		// Mock podman not installed
		mock.SetError("which podman", fmt.Errorf("not found"))

		// Mock podman-compose already installed
		mock.SetOutput("which podman-compose", "/usr/local/bin/podman-compose")

		// Simulate apt lock on first attempt, then success
		attempt := 0
		mock.SetCustomHandler("sudo apt-get", func(cmd string) (string, error) {
			if strings.Contains(cmd, "install -y podman") {
				attempt++
				if attempt == 1 {
					return "", fmt.Errorf("could not get lock /var/lib/dpkg/lock")
				}
				return "", nil
			}
			return "", nil
		})

		err := installer.EnsureInstalled()
		if err != nil {
			t.Errorf("Expected successful installation after retry, got: %v", err)
		}

		if attempt < 2 {
			t.Errorf("Expected at least 2 attempts, got %d", attempt)
		}
	})
}

func TestPodmanInstaller_PodmanComposeInstallation(t *testing.T) {
	t.Run("Install with pip not available", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		// Mock system requirements
		mock.SetOutput("uname -r", "5.4.0")
		mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "7723")
		mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1", "20000000")

		// Mock OS detection
		mock.SetOutput("cat /etc/os-release", "NAME=\"Ubuntu\"")

		// Mock pip and podman-compose installation sequence
		mock.SetError("which podman-compose", fmt.Errorf("not found"))
		mock.SetError("which pip3", fmt.Errorf("not found"))
		
		// Mock successful pip and podman-compose installation
		mock.SetCustomHandler("apt-get", func(cmd string) (string, error) {
			if strings.Contains(cmd, "python3-pip") {
				return "", nil
			}
			return "", fmt.Errorf("unexpected apt-get command")
		})
		mock.SetOutput("sudo pip3 install podman-compose", "")

		err := installer.checkAndInstallPodmanCompose()
		if err != nil {
			t.Errorf("Failed to install podman-compose: %v", err)
		}

		// Verify commands were executed
		commands := mock.GetCommands()
		expectedCommands := []string{
			"which podman-compose",
			"which pip3",
			"python3-pip",
			"sudo pip3 install podman-compose",
		}

		for _, expected := range expectedCommands {
			found := false
			for _, cmd := range commands {
				if strings.Contains(cmd, expected) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected command containing %q not found", expected)
			}
		}
	})
}

func TestPodmanInstaller_SystemRequirements(t *testing.T) {
	mock := NewMockSSHClient()
	installer := NewPodmanInstaller(mock)

	t.Run("Check System Requirements", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		mock.SetOutput("uname -r", "5.4.0")
		mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "7723")
		mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1", "20000000")

		err := installer.checkSystemRequirements()
		if err != nil {
			t.Errorf("System requirements check failed: %v", err)
		}
	})

	t.Run("Insufficient Resources", func(t *testing.T) {
		// Simulate low memory
		mock.SetOutput("free -m", `              total        used        free
Mem:           1024        900        124
Swap:           512        500         12`)

		err := installer.checkSystemRequirements()
		if err == nil {
			t.Error("Expected error for insufficient memory")
		}
	})
}

func TestPodmanInstaller_CleanupOnFailure(t *testing.T) {
	t.Run("Cleanup Partial Installation", func(t *testing.T) {
		mock := NewMockSSHClient()
		installer := NewPodmanInstaller(mock)

		// Mock system requirements
		mock.SetOutput("uname -r", "5.4.0")
		mock.SetOutput("free -m | awk '/^Mem:/{print $2}'", "7723")
		mock.SetOutput("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1", "20000000")

		// Mock OS detection
		mock.SetOutput("cat /etc/os-release", "NAME=\"Ubuntu\"")
		
		// Mock podman not installed
		mock.SetError("which podman", fmt.Errorf("not found"))

		// Mock installation failure and cleanup with a more specific handler
		mock.SetCustomHandler("apt-get", func(cmd string) (string, error) {
			if strings.Contains(cmd, "install") {
				return "", fmt.Errorf("installation failed")
			}
			if strings.Contains(cmd, "remove") || strings.Contains(cmd, "clean") {
				return "", nil // Allow cleanup to succeed
			}
			return "", nil // Allow other apt-get commands to succeed
		})

		err := installer.EnsureInstalled()
		if err == nil {
			t.Error("Expected installation to fail")
		}

		// Verify cleanup was attempted
		commands := mock.GetCommands()
		foundCleanup := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "apt-get remove") {
				foundCleanup = true
				break
			}
		}
		if !foundCleanup {
			t.Error("Cleanup commands not found after failed installation")
		}
	})
}
