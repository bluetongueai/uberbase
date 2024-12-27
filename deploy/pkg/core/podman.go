package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PodmanInstaller struct {
	ssh *SSHConnection
}

func NewPodmanInstaller(ssh *SSHConnection) *PodmanInstaller {
	Logger.Debug("Creating new PodmanInstaller")
	return &PodmanInstaller{
		ssh: ssh,
	}
}

func (p *PodmanInstaller) EnsureInstalled() error {
	Logger.Info("Ensuring Podman is installed")
	if err := p.checkSystemRequirements(); err != nil {
		Logger.Errorf("System requirements not met: %v", err)
		return fmt.Errorf("system requirements not met: %w", err)
	}

	if err := p.checkAndInstallPodman(); err != nil {
		cleanupErr := p.cleanup()
		if cleanupErr != nil {
			Logger.Errorf("Failed to install Podman: %v; cleanup failed: %v", err, cleanupErr)
			return fmt.Errorf("failed to install podman: %w; cleanup failed: %v", err, cleanupErr)
		}
		Logger.Errorf("Failed to install Podman: %v", err)
		return fmt.Errorf("failed to install podman: %w", err)
	}

	if err := p.checkAndInstallPodmanCompose(); err != nil {
		cleanupErr := p.cleanup()
		if cleanupErr != nil {
			Logger.Errorf("Failed to install Podman Compose: %v; cleanup failed: %v", err, cleanupErr)
			return fmt.Errorf("failed to install podman-compose: %w; cleanup failed: %v", err, cleanupErr)
		}
		Logger.Errorf("Failed to install Podman Compose: %v", err)
		return fmt.Errorf("failed to install podman-compose: %w", err)
	}

	Logger.Info("Podman and Podman Compose installed successfully")
	return nil
}

func (p *PodmanInstaller) checkAndInstallPodman() error {
	Logger.Info("Checking if Podman is installed")
	_, err := p.ssh.Exec("which podman")
	if err == nil {
		Logger.Info("Podman is already installed")
		return nil
	}

	// Detect OS and install accordingly
	osType, err := p.detectOS()
	if err != nil {
		Logger.Errorf("Failed to detect OS: %v", err)
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
		Logger.Errorf("Unsupported OS: %s", osType)
		return fmt.Errorf("unsupported OS: %s", osType)
	}

	// Retry logic
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := p.ssh.Exec(installCmd)
		if err != nil {
			if strings.Contains(err.Error(), "could not get lock") && attempt < maxRetries {
				Logger.Warnf("Could not get lock, retrying... (%d/%d)", attempt, maxRetries)
				time.Sleep(2 * time.Second) // Wait before retrying
				continue
			}
			Logger.Errorf("Failed to install Podman: %v", err)
			return fmt.Errorf("failed to install podman: %w", err)
		}
		break
	}

	Logger.Info("Podman installed successfully")
	return nil
}

func (p *PodmanInstaller) checkAndInstallPodmanCompose() error {
	Logger.Info("Checking if Podman Compose is installed")
	// Check if podman-compose is installed
	_, err := p.ssh.Exec("which podman-compose")
	if err == nil {
		Logger.Info("Podman Compose is already installed")
		return nil
	}

	// Install pip if not present
	pipInstallCmd := "which pip3 || ( sudo apt-get update && sudo apt-get install -y python3-pip )"

	_, err = p.ssh.Exec(pipInstallCmd)
	if err != nil {
		Logger.Errorf("Failed to install pip: %v", err)
		return fmt.Errorf("failed to install pip: %w", err)
	}

	// Install podman-compose using pip
	_, err = p.ssh.Exec("sudo pip3 install podman-compose")
	if err != nil {
		Logger.Errorf("Failed to install Podman Compose: %v", err)
		return fmt.Errorf("failed to install podman-compose: %w", err)
	}

	Logger.Info("Podman Compose installed successfully")
	return nil
}

func (p *PodmanInstaller) detectOS() (string, error) {
	Logger.Info("Detecting operating system")
	// Try to detect OS using os-release file
	output, err := p.ssh.Exec("cat /etc/os-release")
	if err != nil {
		Logger.Errorf("Failed to read os-release: %v", err)
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
		Logger.Errorf("Unsupported operating system")
		return "", fmt.Errorf("unsupported operating system")
	}
}

func (p *PodmanInstaller) checkSystemRequirements() error {
	Logger.Info("Checking system requirements")
	// Check memory
	output, err := p.ssh.Exec("free -m | awk '/^Mem:/{print $2}'")
	if err != nil {
		Logger.Errorf("Failed to check memory: %v", err)
		return fmt.Errorf("failed to check memory: %w", err)
	}

	totalMem, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		Logger.Errorf("Failed to parse memory value: %v", err)
		return fmt.Errorf("failed to parse memory value: %w", err)
	}

	if totalMem < 2048 {
		Logger.Errorf("Insufficient memory: %dMB available, minimum required is 2048MB", totalMem)
		return fmt.Errorf("insufficient memory: %dMB available, minimum required is 2048MB", totalMem)
	}

	// Check kernel version (minimum 3.10)
	output, err = p.ssh.Exec("uname -r")
	if err != nil {
		Logger.Errorf("Failed to check kernel version: %v", err)
		return fmt.Errorf("failed to check kernel version: %w", err)
	}

	kernelVersion := strings.TrimSpace(string(output))
	parts := strings.Split(kernelVersion, ".")
	if len(parts) < 2 {
		Logger.Errorf("Invalid kernel version format: %s", kernelVersion)
		return fmt.Errorf("invalid kernel version format: %s", kernelVersion)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		Logger.Errorf("Invalid kernel major version: %v", err)
		return fmt.Errorf("invalid kernel major version: %w", err)
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		Logger.Errorf("Invalid kernel minor version: %v", err)
		return fmt.Errorf("invalid kernel minor version: %w", err)
	}

	if major < 3 || (major == 3 && minor < 10) {
		Logger.Errorf("Kernel version %s is too old, minimum required is 3.10", kernelVersion)
		return fmt.Errorf("kernel version %s is too old, minimum required is 3.10", kernelVersion)
	}

	// Check available disk space (minimum 10GB in /var/lib/containers)
	output, err = p.ssh.Exec("df --output=avail /var/lib/containers 2>/dev/null || df --output=avail / | tail -n 1")
	if err != nil {
		Logger.Errorf("Failed to check disk space: %v", err)
		return fmt.Errorf("failed to check disk space: %w", err)
	}

	availableKB, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		Logger.Errorf("Failed to parse available disk space: %v", err)
		return fmt.Errorf("failed to parse available disk space: %w", err)
	}

	availableGB := availableKB / (1024 * 1024) // Convert KB to GB
	if availableGB < 10 {
		Logger.Errorf("Insufficient disk space: %dGB available, minimum required is 10GB", availableGB)
		return fmt.Errorf("insufficient disk space: %dGB available, minimum required is 10GB", availableGB)
	}

	Logger.Info("System requirements met")
	return nil
}

func (p *PodmanInstaller) cleanup() error {
	Logger.Info("Cleaning up Podman installation")
	osType, err := p.detectOS()
	if err != nil {
		Logger.Errorf("Failed to detect OS for cleanup: %v", err)
		return err
	}

	var cleanupCmd string
	switch osType {
	case "ubuntu", "debian":
		cleanupCmd = "sudo apt-get remove -y podman && sudo apt-get clean"
	case "fedora", "centos", "rhel":
		cleanupCmd = "sudo dnf remove -y podman && sudo dnf clean all"
	default:
		Logger.Errorf("Unsupported OS for cleanup: %s", osType)
		return fmt.Errorf("unsupported OS for cleanup: %s", osType)
	}

	_, err = p.ssh.Exec(cleanupCmd)
	if err != nil {
		Logger.Errorf("Failed to clean up Podman: %v", err)
		return err
	}

	Logger.Info("Podman cleanup successful")
	return nil
}
