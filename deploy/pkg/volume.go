package pkg

import (
	"fmt"
	"os"
	"strings"
)

type VolumeManagerInterface interface {
	EnsureVolume(name string) error
	EnsureVolumes(volumes []string) error
	RemoveVolume(name string) error
	ListVolumes() ([]string, error)
}

type VolumeManager struct {
	ssh SSHClientInterface
}

func NewVolumeManager(ssh SSHClientInterface) *VolumeManager {
	return &VolumeManager{
		ssh: ssh,
	}
}

func (v *VolumeManager) validateVolumeName(name string) error {
	if name == "" {
		return fmt.Errorf("invalid volume name: name cannot be empty")
	}
	// Add more validation as needed
	return nil
}

func (v *VolumeManager) validateVolumeSpec(spec string) error {
	parts := strings.Split(spec, ":")
	if len(parts) > 3 {
		return fmt.Errorf("invalid volume specification: too many parts")
	}

	if len(parts) > 1 {
		// Validate mount options
		if len(parts) == 3 {
			options := strings.Split(parts[2], ",")
			validOptions := map[string]bool{
				"ro": true, "rw": true,
				"z": true, "Z": true,
				"shared": true, "slave": true, "private": true,
				"rshared": true, "rslave": true, "rprivate": true,
				"nocopy": true,
				"consistent": true, "cached": true, "delegated": true,
			}
			
			for _, opt := range options {
				if !validOptions[opt] {
					return fmt.Errorf("invalid mount option: %s", opt)
				}
			}
		}
	}
	return nil
}

func (v *VolumeManager) EnsureVolume(name string) error {
	if err := v.validateVolumeName(name); err != nil {
		return err
	}
	cmd := NewRemoteCommand(v.ssh, fmt.Sprintf("podman volume inspect %s || podman volume create %s", name, name))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to ensure volume %s: %w", name, err)
	}
	return nil
}

func (v *VolumeManager) handleSELinux(hostPath string, options []string) error {
	// Default to private label if Z is specified
	private := false
	shared := false

	for _, opt := range options {
		switch opt {
		case "Z":
			private = true
		case "z":
			shared = true
		}
	}

	if !private && !shared {
		return nil
	}

	// Build semanage command
	var cmd *RemoteCommand
	if private {
		// -t container_file_t for private container content
		cmd = NewRemoteCommand(v.ssh, fmt.Sprintf(
			"chcon -Rt container_file_t %s",
			hostPath,
		))
	} else {
		// -t container_share_t for shared container content
		cmd = NewRemoteCommand(v.ssh, fmt.Sprintf(
			"chcon -Rt container_share_t %s",
			hostPath,
		))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set SELinux context on %s: %w", hostPath, err)
	}

	return nil
}

func (v *VolumeManager) handleVolumeOptions(hostPath string, options []string) []string {
	var mountOpts []string
	var selinuxOpts []string

	for _, opt := range options {
		switch opt {
		case "z", "Z":
			selinuxOpts = append(selinuxOpts, opt)
		case "ro", "rw":
			mountOpts = append(mountOpts, opt)
		case "shared", "slave", "private", "rshared", "rslave", "rprivate", "rbind":
			mountOpts = append(mountOpts, opt)
		case "nocopy":
			mountOpts = append(mountOpts, opt)
		case "consistent", "cached", "delegated":
				mountOpts = append(mountOpts, opt)
		}
	}

	// Handle SELinux separately
	if len(selinuxOpts) > 0 {
		if err := v.handleSELinux(hostPath, selinuxOpts); err != nil {
			fmt.Printf("Warning: SELinux labeling failed: %v\n", err)
		}
		mountOpts = append(mountOpts, selinuxOpts...)
	}

	return mountOpts
}

func (v *VolumeManager) EnsureVolumes(volumes []string) error {
	for _, volume := range volumes {
		if err := v.validateVolumeSpec(volume); err != nil {
			return err
		}

		if strings.Contains(volume, ":") {
			parts := strings.Split(volume, ":")
			hostPath := os.ExpandEnv(parts[0])

			// Handle mount options
			if len(parts) > 2 {
				options := strings.Split(parts[2], ",")
				mountOpts := v.handleVolumeOptions(hostPath, options)
				if len(mountOpts) > 0 {
					// Apply mount options
					// Implementation depends on specific requirements
				}
			}
		} else {
			// Named volume
			if err := v.EnsureVolume(volume); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *VolumeManager) RemoveVolume(name string) error {
	cmd := NewRemoteCommand(v.ssh, fmt.Sprintf("podman volume rm %s", name))
	return cmd.Run()
}

func (v *VolumeManager) ListVolumes() ([]string, error) {
	cmd := NewRemoteCommand(v.ssh, "podman volume ls --format '{{.Name}}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	volumes := strings.Split(strings.TrimSpace(string(output)), "\n")
	return volumes, nil
}
