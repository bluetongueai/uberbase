package pkg

import (
	"fmt"
	"strings"
)

type PodmanInstaller struct {
	ssh *SSHClient
}

func NewPodmanInstaller(ssh *SSHClient) *PodmanInstaller {
	return &PodmanInstaller{
		ssh: ssh,
	}
}

func (p *PodmanInstaller) EnsureInstalled() error {
	if err := p.checkAndInstallPodman(); err != nil {
		return fmt.Errorf("failed to install podman: %w", err)
	}

	if err := p.checkAndInstallPodmanCompose(); err != nil {
		return fmt.Errorf("failed to install podman-compose: %w", err)
	}

	return nil
}

func (p *PodmanInstaller) checkAndInstallPodman() error {
	// Check if podman is installed
	cmd := NewRemoteCommand(p.ssh, "which podman")
	if err := cmd.Run(); err == nil {
		// Podman is already installed
		return nil
	}

	// Detect OS and install accordingly
	osType, err := p.detectOS()
	if err != nil {
		return fmt.Errorf("failed to detect OS: %w", err)
	}

	var installCmd string
	switch osType {
	case "ubuntu", "debian":
		installCmd = strings.Join([]string{
			"sudo apt-get update",
			"sudo apt-get install -y podman",
		}, " && ")
	case "fedora":
		installCmd = "sudo dnf install -y podman"
	case "centos", "rhel":
		installCmd = strings.Join([]string{
			"sudo dnf -y install 'dnf-command(copr)'",
			"sudo dnf -y copr enable rhcontainerbot/container-selinux",
			"sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/CentOS_8/devel:kubic:libcontainers:stable.repo",
			"sudo dnf -y install podman",
		}, " && ")
	default:
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	cmd = NewRemoteCommand(p.ssh, installCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install podman: %w", err)
	}

	return nil
}

func (p *PodmanInstaller) checkAndInstallPodmanCompose() error {
	// Check if podman-compose is installed
	cmd := NewRemoteCommand(p.ssh, "which podman-compose")
	if err := cmd.Run(); err == nil {
		// podman-compose is already installed
		return nil
	}

	// Install pip if not present
	pipInstallCmd := strings.Join([]string{
		"which pip3 || (",
		"sudo apt-get update",
		"sudo apt-get install -y python3-pip",
		")",
	}, " && ")

	cmd = NewRemoteCommand(p.ssh, pipInstallCmd)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install pip: %w", err)
	}

	// Install podman-compose using pip
	cmd = NewRemoteCommand(p.ssh, "sudo pip3 install podman-compose")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to install podman-compose: %w", err)
	}

	return nil
}

func (p *PodmanInstaller) detectOS() (string, error) {
	// Try to detect OS using os-release file
	cmd := NewRemoteCommand(p.ssh, "cat /etc/os-release")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to read os-release: %w", err)
	}

	osRelease := string(output)
	osRelease = strings.ToLower(osRelease)

	switch {
	case strings.Contains(osRelease, "ubuntu"):
		return "ubuntu", nil
	case strings.Contains(osRelease, "debian"):
		return "debian", nil
	case strings.Contains(osRelease, "fedora"):
		return "fedora", nil
	case strings.Contains(osRelease, "centos"):
		return "centos", nil
	case strings.Contains(osRelease, "rhel"):
		return "rhel", nil
	default:
		return "", fmt.Errorf("unsupported operating system")
	}
}
