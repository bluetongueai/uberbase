package deploy

import (
	"fmt"
	"os"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

// Deployer orchestrates the deployment process
type Deployer struct {
	serviceManager *ServiceManager
	trafficManager *TrafficManager
	stateManager   *StateManager
	healthChecker  *HealthChecker
	backupManager  *BackupManager
	metrics        *DeploymentMetrics
	txManager      *TransactionManager
	ssh            *core.SSHConnection
}

func NewDeployer(ssh *core.SSHConnection, workDir string, registryConfig podman.RegistryConfig) *Deployer {
	metrics := &DeploymentMetrics{}
	stateManager := NewStateManager(ssh, workDir)
	return &Deployer{
		serviceManager: NewServiceManager(ssh, registryConfig),
		trafficManager: NewTrafficManager(ssh),
		stateManager:   stateManager,
		healthChecker:  NewHealthChecker(ssh, metrics),
		backupManager:  NewBackupManager(ssh),
		metrics:        metrics,
		txManager:      NewTransactionManager(stateManager),
		ssh:            ssh,
	}
}

func (d *Deployer) Deploy(service *Service, newVersion string) error {
	// Acquire deployment lock
	hostname, _ := os.Hostname()
	if err := d.stateManager.AcquireLock(hostname); err != nil {
		return fmt.Errorf("failed to acquire deployment lock: %w", err)
	}
	defer d.stateManager.ReleaseLock(hostname)

	state, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	currentState := state.Services[service.Name]

	// If this is not the first deployment, prepare rollback info
	if currentState != nil && currentState.BlueVersion != "" {
		rollbackInfo := &RollbackInfo{
			PreviousVersion: currentState.BlueVersion,
			VolumeBackups:   make(map[string]string),
			Timestamp:       time.Now().UTC().Format(time.RFC3339),
		}

		// Backup volumes before migration
		for _, volume := range service.Volumes {
			if volume.Persistent {
				backupPath := fmt.Sprintf("/var/lib/podman/volumes/backups/%s-%s-%s",
					service.Name, currentState.BlueVersion, volume.Name)

				if err := d.serviceManager.volumeManager.BackupVolume(volume.Name, backupPath); err != nil {
					return fmt.Errorf("failed to backup volume %s: %w", volume.Name, err)
				}
				rollbackInfo.VolumeBackups[volume.Name] = backupPath
			}
		}

		currentState.RollbackInfo = rollbackInfo
		if err := d.stateManager.Save(state); err != nil {
			return fmt.Errorf("failed to save rollback info: %w", err)
		}
	}

	// If this is not the first deployment, handle volume migration
	if currentState != nil && currentState.BlueVersion != "" {
		if err := d.serviceManager.migrateVolumes(service, currentState.BlueVersion, newVersion); err != nil {
			return fmt.Errorf("volume migration failed: %w", err)
		}
	}

	// Create new version (green)
	if err := d.serviceManager.CreateService(service, newVersion); err != nil {
		return fmt.Errorf("failed to create new version: %w", err)
	}

	// Handle backup configuration
	if err := d.handleBackups(service); err != nil {
		core.Logger.Warnf("Failed to configure backups: %v", err)
	}

	// If this is the first deployment, we're done
	if currentState == nil || currentState.BlueVersion == "" {
		state.Services[service.Name] = &ServiceState{
			GreenVersion: newVersion,
			GreenWeight:  100,
			LastUpdated:  time.Now().UTC().Format(time.RFC3339),
		}
		return d.stateManager.Save(state)
	}

	// Wait for health check if configured
	if service.Container.Healthcheck != nil {
		if err := d.waitForHealthy(service, newVersion); err != nil {
			// Rollback on health check failure
			_ = d.serviceManager.RemoveService(service, newVersion)
			return fmt.Errorf("health check failed: %w", err)
		}
	}

	// Gradually shift traffic
	if err := d.trafficManager.UpdateTraffic(service, currentState.BlueVersion, newVersion, 90, 10); err != nil {
		return fmt.Errorf("failed initial traffic shift: %w", err)
	}

	// Update state for initial shift
	currentState.GreenVersion = newVersion
	currentState.BlueWeight = 90
	currentState.GreenWeight = 10
	if err := d.stateManager.Save(state); err != nil {
		return fmt.Errorf("failed to save state: %w", err)
	}

	// Complete the transition
	if err := d.trafficManager.UpdateTraffic(service, currentState.BlueVersion, newVersion, 0, 100); err != nil {
		return fmt.Errorf("failed final traffic shift: %w", err)
	}

	// Remove old version
	if err := d.serviceManager.RemoveService(service, currentState.BlueVersion); err != nil {
		core.Logger.Warnf("Failed to remove old version: %v", err)
	}

	// Update final state
	currentState.BlueVersion = newVersion
	currentState.GreenVersion = ""
	currentState.BlueWeight = 100
	currentState.GreenWeight = 0
	currentState.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	return d.stateManager.Save(state)
}

func (d *Deployer) waitForHealthy(service *Service, version string) error {
	containerName := fmt.Sprintf("%s-%s", service.Name, version)
	maxAttempts := 30 // 5 minutes with 10s sleep
	attempt := 0

	for attempt < maxAttempts {
		output, err := d.serviceManager.containerMgr.Inspect(containerName)
		if err != nil {
			return fmt.Errorf("failed to inspect container: %w", err)
		}

		if output.Health.Status == "healthy" {
			return nil
		}

		attempt++
		time.Sleep(10 * time.Second)
	}

	return fmt.Errorf("container failed to become healthy within timeout")
}

// Add new method to deploy all services
func (d *Deployer) DeployCompose(config *ComposeConfig, version string) error {
	_, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Convert all compose services to our Service type
	services := make(map[string]*Service)
	for name, cs := range config.Services {
		service, err := ConvertComposeService(name, cs)
		if err != nil {
			return fmt.Errorf("failed to convert service %s: %w", name, err)
		}
		services[name] = service
	}

	// Build dependency graph and get deployment order
	order, err := buildDeploymentOrder(services)
	if err != nil {
		return fmt.Errorf("failed to build deployment order: %w", err)
	}

	// Deploy services in order
	for _, serviceName := range order {
		service := services[serviceName]
		if err := d.Deploy(service, version); err != nil {
			d.rollbackServices(services, serviceName)
			return fmt.Errorf("failed to deploy service %s: %w", serviceName, err)
		}
	}

	return nil
}

// Helper function to rollback services in case of failure
func (d *Deployer) rollbackServices(services map[string]*Service, failedService string) {
	for name, service := range services {
		if name == failedService {
			break // Stop at failed service
		}
		// Attempt to restore previous version
		state, err := d.stateManager.Load()
		if err != nil {
			core.Logger.Errorf("Failed to load state during rollback: %v", err)
			continue
		}

		if currentState := state.Services[name]; currentState != nil && currentState.BlueVersion != "" {
			if err := d.Deploy(service, currentState.BlueVersion); err != nil {
				core.Logger.Errorf("Failed to rollback service %s: %v", name, err)
			}
		}
	}
}

// buildDeploymentOrder returns a slice of service names in deployment order
func buildDeploymentOrder(services map[string]*Service) ([]string, error) {
	// Implementation of topological sort based on depends_on
	// This is a simplified version - you might want to add cycle detection
	var order []string
	visited := make(map[string]bool)

	var visit func(name string) error
	visit = func(name string) error {
		if visited[name] {
			return nil
		}
		visited[name] = true

		service := services[name]
		if service.Container != nil {
			for depName := range service.Container.DependsOn {
				if err := visit(depName); err != nil {
					return err
				}
			}
		}
		order = append(order, name)
		return nil
	}

	for name := range services {
		if err := visit(name); err != nil {
			return nil, err
		}
	}

	return order, nil
}

// Add backup handling method
func (d *Deployer) handleBackups(service *Service) error {
	if service.Container == nil || service.Container.Labels == nil {
		return nil
	}

	// Check if backups are enabled via labels
	if enabled, ok := service.Container.Labels[BackupEnabled]; !ok || enabled != "true" {
		return nil
	}

	// Store backup configuration in service state
	state, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state for backup config: %w", err)
	}

	if state.Services[service.Name] == nil {
		state.Services[service.Name] = &ServiceState{}
	}

	// Update state with backup information
	currentState := state.Services[service.Name]
	if currentState.VolumeStates == nil {
		currentState.VolumeStates = make(map[string]*VolumeState)
	}

	// Update volume states with backup info
	for _, vol := range service.Volumes {
		if vol.Persistent {
			currentState.VolumeStates[vol.Name] = &VolumeState{
				Name: vol.Name,
			}
		}
	}

	return d.stateManager.Save(state)
}

// Add rollback method
func (d *Deployer) Rollback(service *Service) error {
	// Acquire deployment lock
	hostname, _ := os.Hostname()
	if err := d.stateManager.AcquireLock(hostname); err != nil {
		return fmt.Errorf("failed to acquire deployment lock: %w", err)
	}
	defer d.stateManager.ReleaseLock(hostname)

	state, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	currentState := state.Services[service.Name]
	if currentState == nil || currentState.RollbackInfo == nil {
		return fmt.Errorf("no rollback information available for service %s", service.Name)
	}

	// Restore volumes from backups
	for volName, backupPath := range currentState.RollbackInfo.VolumeBackups {
		if err := d.serviceManager.volumeManager.RestoreVolume(volName, backupPath); err != nil {
			return fmt.Errorf("failed to restore volume %s: %w", volName, err)
		}
	}

	// Deploy previous version
	if err := d.Deploy(service, currentState.RollbackInfo.PreviousVersion); err != nil {
		return fmt.Errorf("failed to rollback to version %s: %w",
			currentState.RollbackInfo.PreviousVersion, err)
	}

	return nil
}

// Add this method to the Deployer struct
func (d *Deployer) StateManager() *StateManager {
	return d.stateManager
}

// Add method to get metrics
func (d *Deployer) Metrics() *DeploymentMetrics {
	return d.metrics
}

func (d *Deployer) ResumeDeployment(services map[string]*Service, version string) error {
	// Validate all services first
	for _, service := range services {
		if err := service.Validate(); err != nil {
			return fmt.Errorf("service validation failed: %w", err)
		}
	}

	// Get deployment order
	order, err := buildDeploymentOrder(services)
	if err != nil {
		return fmt.Errorf("failed to determine deployment order: %w", err)
	}

	// Find last successful deployment
	var lastSuccessful string
	for _, name := range order {
		tx, err := d.txManager.GetLastTransaction(name)
		if err != nil {
			return fmt.Errorf("failed to get transaction history: %w", err)
		}
		if tx != nil && tx.Status == TxCompleted {
			lastSuccessful = name
		}
	}

	// Resume from last successful deployment
	resumeFound := false
	for _, name := range order {
		if !resumeFound {
			if name == lastSuccessful {
				resumeFound = true
			}
			continue
		}

		service := services[name]
		tx := TransactionLog{
			ServiceName: name,
			Action:      "deploy",
			Status:      TxStarted,
			Timestamp:   time.Now(),
			Version:     version,
		}

		if err := d.txManager.LogTransaction(tx); err != nil {
			return fmt.Errorf("failed to log transaction: %w", err)
		}

		if err := d.Deploy(service, version); err != nil {
			tx.Status = TxFailed
			tx.Error = err.Error()
			d.txManager.LogTransaction(tx)
			return fmt.Errorf("failed to deploy service %s: %w", name, err)
		}

		tx.Status = TxCompleted
		if err := d.txManager.LogTransaction(tx); err != nil {
			return fmt.Errorf("failed to log transaction completion: %w", err)
		}
	}

	return nil
}

// Add this method to the Deployer struct
func (d *Deployer) Host() string {
	return d.ssh.Host()
}

// RollbackCompose rolls back all services in a compose configuration to their previous versions
func (d *Deployer) RollbackCompose(config *ComposeConfig) error {
	// Acquire deployment lock
	hostname, _ := os.Hostname()
	if err := d.stateManager.AcquireLock(hostname); err != nil {
		return fmt.Errorf("failed to acquire deployment lock: %w", err)
	}
	defer d.stateManager.ReleaseLock(hostname)

	// Load current state
	state, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Convert compose services to our Service type
	services := make(map[string]*Service)
	for name, cs := range config.Services {
		service, err := ConvertComposeService(name, cs)
		if err != nil {
			return fmt.Errorf("failed to convert service %s: %w", name, err)
		}
		services[name] = service
	}

	// Get deployment order in reverse (for rollback)
	order, err := buildDeploymentOrder(services)
	if err != nil {
		return fmt.Errorf("failed to build deployment order: %w", err)
	}

	// Reverse the order for rollback
	for i := 0; i < len(order)/2; i++ {
		order[i], order[len(order)-1-i] = order[len(order)-1-i], order[i]
	}

	// Roll back each service in reverse order
	for _, serviceName := range order {
		service := services[serviceName]
		currentState := state.Services[serviceName]

		if currentState == nil || currentState.RollbackInfo == nil {
			core.Logger.Warnf("No rollback information available for service %s, skipping", serviceName)
			continue
		}

		if err := d.Rollback(service); err != nil {
			return fmt.Errorf("failed to rollback service %s: %w", serviceName, err)
		}
	}

	return nil
}
