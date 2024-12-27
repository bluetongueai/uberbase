package deploy

import (
	"fmt"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

// ServiceManager handles the lifecycle of services (containers/pods)
type ServiceManager struct {
	podManager     *podman.PodManager
	containerMgr   *podman.ContainerManager
	networkManager *podman.NetworkManager
	volumeManager  *podman.VolumeManager
	imageManager   *podman.ImageManager
}

func NewServiceManager(ssh *core.SSHConnection, registryConfig podman.RegistryConfig) *ServiceManager {
	return &ServiceManager{
		podManager:     podman.NewPodManager(ssh),
		containerMgr:   podman.NewContainerManager(ssh),
		networkManager: podman.NewNetworkManager(ssh),
		volumeManager:  podman.NewVolumeManager(ssh),
		imageManager:   podman.NewImageManager(ssh, registryConfig),
	}
}

// CreateService creates a new service instance
func (s *ServiceManager) CreateService(service *Service, version string) error {
	containerName := fmt.Sprintf("%s-%s", service.Name, version)

	// Setup infrastructure
	if err := s.setupInfrastructure(service); err != nil {
		return fmt.Errorf("infrastructure setup failed: %w", err)
	}

	// Ensure image is available
	if err := s.imageManager.EnsureImage(service.Image); err != nil {
		return fmt.Errorf("failed to ensure image: %w", err)
	}

	// Create container configuration
	container := service.Container
	container.Name = containerName
	container.Image = service.Image.String()

	// Create the container
	if err := s.containerMgr.Create(*container); err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	return nil
}

func (s *ServiceManager) RemoveService(service *Service, version string) error {
	containerName := fmt.Sprintf("%s-%s", service.Name, version)

	// Remove container
	if err := s.containerMgr.Remove(containerName); err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}

	// Cleanup infrastructure if needed
	return s.cleanupInfrastructure(service)
}

func (s *ServiceManager) setupInfrastructure(service *Service) error {
	// Create volumes
	for _, volume := range service.Volumes {
		if err := s.volumeManager.EnsureVolume(volume.Name); err != nil {
			return fmt.Errorf("failed to create volume: %w", err)
		}
	}

	// Create network
	if service.Network != nil {
		if err := s.networkManager.EnsureNetwork(service.Network.Name, service.Network.Internal); err != nil {
			return fmt.Errorf("failed to create network: %w", err)
		}
	}

	return nil
}

func (s *ServiceManager) cleanupInfrastructure(service *Service) error {
	// Only cleanup volumes that aren't persistent
	for _, volume := range service.Volumes {
		if !volume.Persistent {
			if err := s.volumeManager.RemoveVolume(volume.Name); err != nil {
				core.Logger.Warnf("Failed to remove volume %s: %v", volume.Name, err)
			}
		}
	}

	// Network cleanup could be added here if needed
	return nil
}

func (s *ServiceManager) migrateVolumes(service *Service, oldVersion, newVersion string) error {
	if len(service.Volumes) == 0 {
		return nil
	}

	core.Logger.Info("Starting volume migration")

	// Get volume names from service volumes
	var volumeNames []string
	for _, volume := range service.Volumes {
		if volume.Persistent {
			volumeNames = append(volumeNames, volume.Name)
		}
	}

	// Use container manager to handle migration
	if err := s.containerMgr.MigrateContainer(service.Name, oldVersion, newVersion, volumeNames); err != nil {
		return fmt.Errorf("container migration failed: %w", err)
	}

	return nil
}
