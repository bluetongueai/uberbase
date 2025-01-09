package core

import (
	"fmt"
	"strings"
)

type Platform struct {
	OS      string // Darwin, Linux, Windows
	Distro  string // debian, rhel, alpine, etc.
	HasSudo bool
}

type Installer struct {
	executor Executor
	platform *Platform
}

func NewInstaller(executor Executor) *Installer {
	return &Installer{
		executor: executor,
	}
}

// DetectPlatform identifies the operating system and distribution
func (i *Installer) DetectPlatform() error {
	os, err := i.executor.Exec("uname -s")
	if err != nil {
		return fmt.Errorf("failed to detect OS: %w", err)
	}

	platform := &Platform{
		OS: strings.TrimSpace(string(os)),
	}

	// Check sudo availability
	_, err = i.executor.Exec("sudo -n true")
	platform.HasSudo = err == nil

	// Detect Linux distribution
	if platform.OS == "Linux" {
		// Try to detect distribution using os-release
		if output, err := i.executor.Exec("cat /etc/os-release"); err == nil {
			if strings.Contains(output, "ID=debian") || strings.Contains(output, "ID=ubuntu") {
				platform.Distro = "debian"
			} else if strings.Contains(output, "ID=rhel") || strings.Contains(output, "ID=fedora") {
				platform.Distro = "rhel"
			} else if strings.Contains(output, "ID=alpine") {
				platform.Distro = "alpine"
			}
		}
	}

	i.platform = platform
	return nil
}

// Capability checks
func (i *Installer) HasGit() bool {
	_, err := i.executor.Exec("git --version")
	return err == nil
}

func (i *Installer) HasDocker() bool {
	_, err := i.executor.Exec("docker --version")
	return err == nil
}

func (i *Installer) HasDockerCompose() bool {
	_, err := i.executor.Exec("docker compose --version")
	return err == nil
}

func (i *Installer) HasPodman() bool {
	_, err := i.executor.Exec("podman --version")
	return err == nil
}

func (i *Installer) HasPodmanCompose() bool {
	_, err := i.executor.Exec("podman-compose --version")
	return err == nil
}

// Installation functions
func (i *Installer) InstallGit() error {
	if i.platform == nil {
		if err := i.DetectPlatform(); err != nil {
			return err
		}
	}

	switch i.platform.OS {
	case "Darwin":
		_, err := i.executor.Exec("brew install git")
		return err
	case "Linux":
		switch i.platform.Distro {
		case "debian":
			cmd := "apt-get update && apt-get install -y git"
			if i.platform.HasSudo {
				cmd = "sudo " + cmd
			}
			_, err := i.executor.Exec(cmd)
			return err
		case "rhel":
			cmd := "dnf install -y git"
			if i.platform.HasSudo {
				cmd = "sudo " + cmd
			}
			_, err := i.executor.Exec(cmd)
			return err
		case "alpine":
			cmd := "apk add git"
			if i.platform.HasSudo {
				cmd = "sudo " + cmd
			}
			_, err := i.executor.Exec(cmd)
			return err
		}
	case "Windows":
		_, err := i.executor.Exec("winget install -e --id Git.Git")
		return err
	}
	return fmt.Errorf("unsupported platform: %s (%s)", i.platform.OS, i.platform.Distro)
}

func (i *Installer) InstallContainerRuntime() error {
	if i.platform == nil {
		if err := i.DetectPlatform(); err != nil {
			return err
		}
	}

	// Try Podman first
	if err := i.installPodman(); err == nil {
		return i.installPodmanCompose()
	}

	// Fall back to Docker if Podman installation fails
	return i.installDocker()
}

func (i *Installer) installPodman() error {
	switch i.platform.OS {
	case "Darwin":
		_, err := i.executor.Exec("brew install podman")
		return err
	case "Linux":
		switch i.platform.Distro {
		case "debian":
			cmd := "apt-get update && apt-get install -y podman"
			if i.platform.HasSudo {
				cmd = "sudo " + cmd
			}
			_, err := i.executor.Exec(cmd)
			return err
		case "rhel":
			cmd := "dnf install -y podman"
			if i.platform.HasSudo {
				cmd = "sudo " + cmd
			}
			_, err := i.executor.Exec(cmd)
			return err
		}
	case "Windows":
		_, err := i.executor.Exec("winget install -e --id RedHat.Podman")
		return err
	}
	return fmt.Errorf("podman installation not supported on %s (%s)", i.platform.OS, i.platform.Distro)
}

func (i *Installer) installPodmanCompose() error {
	switch i.platform.OS {
	case "Darwin":
		_, err := i.executor.Exec("brew install podman-compose")
		return err
	case "Linux":
		cmd := "pip3 install podman-compose"
		if i.platform.HasSudo {
			cmd = "sudo " + cmd
		}
		_, err := i.executor.Exec(cmd)
		return err
	}
	return fmt.Errorf("podman-compose installation not supported on %s", i.platform.OS)
}

func (i *Installer) installDocker() error {
	switch i.platform.OS {
	case "Darwin":
		_, err := i.executor.Exec("brew install --cask docker")
		return err
	case "Linux":
		// Use Docker's official installation script
		getDocker := "curl -fsSL https://get.docker.com -o get-docker.sh"
		if _, err := i.executor.Exec(getDocker); err != nil {
			return err
		}

		cmd := "sh get-docker.sh"
		if i.platform.HasSudo {
			cmd = "sudo " + cmd
		}
		_, err := i.executor.Exec(cmd)
		return err
	case "Windows":
		_, err := i.executor.Exec("winget install -e --id Docker.DockerDesktop")
		return err
	}
	return fmt.Errorf("docker installation not supported on %s", i.platform.OS)
}
