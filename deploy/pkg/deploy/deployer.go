package deploy

import (
	"fmt"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

// Deployer orchestrates the deployment process
type Deployer struct {
	serviceManager *ServiceManager
	trafficManager *TrafficManager
	stateManager   *StateManager
}

func NewDeployer(ssh *core.SSHConnection, workDir string, registryConfig podman.RegistryConfig) *Deployer {
	return &Deployer{
		serviceManager: NewServiceManager(ssh, registryConfig),
		trafficManager: NewTrafficManager(ssh),
		stateManager:   NewStateManager(ssh, workDir),
	}
}

func (d *Deployer) Deploy(service *Service, newVersion string) error {
	state, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load state: %w", err)
	}

	// Create new version (green)
	if err := d.serviceManager.CreateService(service, newVersion); err != nil {
		return fmt.Errorf("failed to create new version: %w", err)
	}

	// If this is the first deployment, we're done
	currentState := state.Services[service.Name]
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
