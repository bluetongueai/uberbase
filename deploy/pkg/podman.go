package pkg

import (
	"fmt"
	"strconv"
	"strings"
)

type PodmanInstaller struct {
	ssh SSHClientInterface
}

func NewPodmanInstaller(ssh SSHClientInterface) *PodmanInstaller {
	return &PodmanInstaller{
		ssh: ssh,
	}
}

func (p *PodmanInstaller) EnsureInstalled() error {
	if err := p.checkSystemRequirements(); err != nil {
		return fmt.Errorf("system requirements not met: %w", err)
	}

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
		installCmd = "sudo apt-get update && sudo apt-get install -y podman"
	case "fedora":
		installCmd = "sudo dnf install -y podman"
	case "centos", "rhel":
		installCmd = "sudo dnf -y install 'dnf-command(copr)' && " +
			"sudo dnf -y copr enable rhcontainerbot/container-selinux && " +
			"sudo curl -L -o /etc/yum.repos.d/devel:kubic:libcontainers:stable.repo " +
			"https://download.opensuse.org/repositories/devel:/kubic:/libcontainers:/stable/CentOS_8/devel:kubic:libcontainers:stable.repo && " +
			"sudo dnf -y install podman"
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
	pipInstallCmd := "which pip3 || ( sudo apt-get update && sudo apt-get install -y python3-pip )"

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

func (p *PodmanInstaller) checkSystemRequirements() error {
	// Check memory
	cmd := NewRemoteCommand(p.ssh, "free -m | awk '/^Mem:/{print $2}'")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check memory: %w", err)
	}
	
	totalMem, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse memory value: %w", err)
	}
	
	if totalMem < 2048 {
		return fmt.Errorf("insufficient memory: %dMB available, minimum required is 2048MB", totalMem)
	}

	// Check kernel version (minimum 3.10)
	cmd = NewRemoteCommand(p.ssh, "uname -r")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check kernel version: %w", err)
	}
	
	kernelVersion := strings.TrimSpace(string(output))
	parts := strings.Split(kernelVersion, ".")
	if len(parts) < 2 {
		return fmt.Errorf("invalid kernel version format: %s", kernelVersion)
	}
	
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return fmt.Errorf("invalid kernel major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("invalid kernel minor version: %w", err)
	}
	
	if major < 3 || (major == 3 && minor < 10) {
		return fmt.Errorf("kernel version %s is too old, minimum required is 3.10", kernelVersion)
	}

	// Check available disk space (minimum 10GB in /var/lib/containers)
	cmd = NewRemoteCommand(p.ssh, "df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1")
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check disk space: %w", err)
	}
	
	availableKB, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return fmt.Errorf("failed to parse available disk space: %w", err)
	}
	
	availableGB := availableKB / (1024 * 1024) // Convert KB to GB
	if availableGB < 10 {
		return fmt.Errorf("insufficient disk space: %dGB available, minimum required is 10GB", availableGB)
	}

	return nil
}

func (p *PodmanInstaller) cleanup() error {
	osType, err := p.detectOS()
	if err != nil {
		return err
	}

	var cleanupCmd string
	switch osType {
	case "ubuntu", "debian":
		cleanupCmd = "sudo apt-get remove -y podman && sudo apt-get clean"
	case "fedora", "centos", "rhel":
		cleanupCmd = "sudo dnf remove -y podman && sudo dnf clean all"
	default:
		return fmt.Errorf("unsupported OS for cleanup: %s", osType)
	}

	cmd := NewRemoteCommand(p.ssh, cleanupCmd)
	return cmd.Run()
}
