package deploy

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type DeploymentCoordinator struct {
	deployers map[string]*Deployer // key is hostname
	mu        sync.Mutex
}

func NewDeploymentCoordinator(deployers []*Deployer) *DeploymentCoordinator {
	deployerMap := make(map[string]*Deployer)
	for _, d := range deployers {
		deployerMap[d.Host()] = d
	}

	return &DeploymentCoordinator{
		deployers: deployerMap,
	}
}

func (c *DeploymentCoordinator) getHostsForService(service *Service) []string {
	if len(service.Placement.Constraints) == 0 {
		// No placement constraints - use all hosts
		hosts := make([]string, 0, len(c.deployers))
		for host := range c.deployers {
			hosts = append(hosts, host)
		}
		return hosts
	}

	// Get hosts from placement constraints
	hosts := make([]string, 0)
	for _, constraint := range service.Placement.Constraints {
		if constraint.Type == "host" && constraint.Operation == "=" {
			hosts = append(hosts, constraint.Value)
		}
	}
	return hosts
}

func (c *DeploymentCoordinator) validateHostAvailability(config *ComposeConfig) error {
	requiredHosts := make(map[string][]string) // host -> services requiring it

	// Collect all required hosts from service placement labels
	for serviceName, cs := range config.Services {
		if cs.Labels == nil {
			continue
		}

		// Check single host
		if host := cs.Labels[PlacementHost]; host != "" {
			if _, ok := c.deployers[host]; !ok {
				return fmt.Errorf("service %s requires host %s which was not provided to deploy tool", serviceName, host)
			}
			requiredHosts[host] = append(requiredHosts[host], serviceName)
		}

		// Check multiple hosts
		if hosts := cs.Labels[PlacementHosts]; hosts != "" {
			for _, host := range strings.Split(hosts, ",") {
				host = strings.TrimSpace(host)
				if _, ok := c.deployers[host]; !ok {
					return fmt.Errorf("service %s requires host %s which was not provided to deploy tool", serviceName, host)
				}
				requiredHosts[host] = append(requiredHosts[host], serviceName)
			}
		}
	}

	return nil
}

func (c *DeploymentCoordinator) DeployCompose(config *ComposeConfig, version string) error {
	// Validate all required hosts are available before starting deployment
	if err := c.validateHostAvailability(config); err != nil {
		return err
	}

	// Convert services and build deployment plan
	services := make(map[string]*Service)
	deploymentPlan := make(map[string][]string) // service name -> host list

	for name, cs := range config.Services {
		service, err := ConvertComposeService(name, cs)
		if err != nil {
			return fmt.Errorf("failed to convert service %s: %w", name, err)
		}
		services[name] = service

		// If no placement constraints are specified, use all available hosts
		if len(service.Placement.Constraints) == 0 {
			// Get list of all hosts from deployers map
			hosts := make([]string, 0, len(c.deployers))
			for host := range c.deployers {
				hosts = append(hosts, host)
			}
			deploymentPlan[name] = hosts
		} else {
			// Use placement constraints to determine hosts
			hosts := c.getHostsForService(service)
			deploymentPlan[name] = hosts
		}
	}

	// Execute deployment plan
	order, err := buildDeploymentOrder(services)
	if err != nil {
		return fmt.Errorf("failed to build deployment order: %w", err)
	}

	for _, serviceName := range order {
		service := services[serviceName]
		targetHosts := deploymentPlan[serviceName]

		// Deploy to selected hosts in parallel
		errChan := make(chan error, len(targetHosts))
		var wg sync.WaitGroup

		for _, hostname := range targetHosts {
			deployer := c.deployers[hostname]
			wg.Add(1)
			go func(hostname string, deployer *Deployer) {
				defer wg.Done()
				if err := deployer.Deploy(service, version); err != nil {
					errChan <- fmt.Errorf("deployment failed on %s: %w", hostname, err)
				}
			}(hostname, deployer)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			return err
		}
	}

	return nil
}

func (c *DeploymentCoordinator) Rollback(config *ComposeConfig) error {
	// Acquire locks on all hosts
	if err := c.acquireAllLocks(); err != nil {
		return fmt.Errorf("failed to acquire locks: %w", err)
	}
	defer c.releaseAllLocks()

	var errors []error
	var wg sync.WaitGroup

	for _, d := range c.deployers {
		wg.Add(1)
		go func(deployer *Deployer) {
			defer wg.Done()
			if err := deployer.RollbackCompose(config); err != nil {
				c.mu.Lock()
				errors = append(errors, fmt.Errorf("rollback failed on %s: %w", deployer.Host(), err))
				c.mu.Unlock()
			}
		}(d)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("rollback failed on some hosts: %v", errors)
	}

	return nil
}

func (c *DeploymentCoordinator) RollbackService(service *Service) error {
	// Acquire locks on all hosts
	if err := c.acquireAllLocks(); err != nil {
		return fmt.Errorf("failed to acquire locks: %w", err)
	}
	defer c.releaseAllLocks()

	var errors []error
	var wg sync.WaitGroup

	for _, d := range c.deployers {
		wg.Add(1)
		go func(deployer *Deployer) {
			defer wg.Done()
			if err := deployer.Rollback(service); err != nil {
				c.mu.Lock()
				errors = append(errors, fmt.Errorf("rollback failed on %s: %w", deployer.Host(), err))
				c.mu.Unlock()
			}
		}(d)
	}

	wg.Wait()

	if len(errors) > 0 {
		return fmt.Errorf("rollback failed on some hosts: %v", errors)
	}

	return nil
}

func (c *DeploymentCoordinator) acquireAllLocks() error {
	var errors []error
	var wg sync.WaitGroup

	for _, d := range c.deployers {
		wg.Add(1)
		go func(deployer *Deployer) {
			defer wg.Done()
			if err := deployer.StateManager().AcquireLock(deployer.Host()); err != nil {
				c.mu.Lock()
				errors = append(errors, fmt.Errorf("failed to acquire lock on %s: %w", deployer.Host(), err))
				c.mu.Unlock()
			}
		}(d)
	}

	wg.Wait()

	if len(errors) > 0 {
		// Release any acquired locks
		c.releaseAllLocks()
		return fmt.Errorf("failed to acquire all locks: %v", errors)
	}

	return nil
}

func (c *DeploymentCoordinator) releaseAllLocks() {
	for _, d := range c.deployers {
		if err := d.StateManager().ReleaseLock(d.Host()); err != nil {
			// Just log the error since we're cleaning up
			core.Logger.Errorf("Failed to release lock on %s: %v", d.Host(), err)
		}
	}
}
