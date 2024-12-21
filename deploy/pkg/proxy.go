package pkg

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type ProxyService struct {
	Name        string
	Image       string
	Domains     []string
	SSL         bool
	Networks    []string
	Environment map[string]string
	Volumes     []string
	Command     []string
	Private     bool
	Port        string
	Version     string
	Labels      map[string]string
	Weight      int
	HealthCheckTimeout time.Duration
}

type ProxyManager struct {
	ssh                SSHClientInterface
	proxyBin           string
	containerMgr       *ContainerManager
	healthCheckTimeout time.Duration
	healthCheckInterval time.Duration
	ctx                context.Context
}

func NewProxyManager(ssh SSHClientInterface, proxyBin string) *ProxyManager {
	return &ProxyManager{
		ssh:                ssh,
		proxyBin:           proxyBin,
		containerMgr:       NewContainerManager(ssh),
		healthCheckTimeout: 5 * time.Second,
		healthCheckInterval: 100 * time.Millisecond,
		ctx:                context.Background(),
	}
}

func (p *ProxyManager) DeployService(service ProxyService) error {
	if err := p.validateService(service); err != nil {
		return fmt.Errorf("invalid service configuration: %w", err)
	}
	containerName := fmt.Sprintf("%s-%s", service.Name, service.Version)
	
	labels := map[string]string{
		"traefik.enable": "true",
	}
	
	for _, domain := range service.Domains {
		routerName := strings.ReplaceAll(domain, ".", "-")
		labels[fmt.Sprintf("traefik.http.routers.%s.rule", routerName)] = fmt.Sprintf("Host(`%s`)", domain)
		
		if service.SSL {
			labels[fmt.Sprintf("traefik.http.routers.%s.tls", routerName)] = "true"
			labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", routerName)] = "default"
		}
	}

	labels["traefik.http.services."+service.Name+".loadbalancer.server.port"] = service.Port
	
	if service.Weight > 0 {
		labels["traefik.http.services."+service.Name+".loadbalancer.server.weight"] = fmt.Sprintf("%d", service.Weight)
	}

	for k, v := range service.Labels {
		labels[k] = v
	}

	cmd := NewRemoteCommand(p.ssh, fmt.Sprintf(
		"podman run -d --name %s %s %s %s %s %s",
		containerName,
		p.buildLabelsArg(labels),
		p.buildEnvArg(service.Environment),
		p.buildNetworksArg(service.Networks),
		service.Image,
		strings.Join(service.Command, " "),
	))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	if service.HealthCheckTimeout > 0 {
		if err := p.waitForHealthyContainer(containerName, service.HealthCheckTimeout); err != nil {
			removeErr := p.RemoveVersion(service.Name, service.Version)
			if removeErr != nil {
				return fmt.Errorf("container unhealthy and cleanup failed: %v (original error: %w)", removeErr, err)
			}
			return fmt.Errorf("container failed health check: %w", err)
		}
	}

	return nil
}

func (p *ProxyManager) SwitchTraffic(serviceName, blueVersion, greenVersion string, blueWeight, greenWeight int) error {
	blueContainer := fmt.Sprintf("%s-%s", serviceName, blueVersion)
	greenContainer := fmt.Sprintf("%s-%s", serviceName, greenVersion)

	// Check if blue version is healthy
	if err := p.waitForHealthy(blueContainer); err != nil {
		return fmt.Errorf("blue version not healthy: %w", err)
	}

	// Update green version weight first
	if err := p.updateWeight(serviceName, greenVersion, greenWeight); err != nil {
		return fmt.Errorf("failed to update green version weight: %w", err)
	}

	// Check if green version is healthy
	if err := p.waitForHealthy(greenContainer); err != nil {
		// Rollback on failure
		if rollbackErr := p.updateWeight(serviceName, greenVersion, 0); rollbackErr != nil {
			return fmt.Errorf("green version unhealthy and rollback failed: %v (original error: %w)", rollbackErr, err)
		}
		return fmt.Errorf("green version not healthy: %w", err)
	}

	// Update blue version weight
	if err := p.updateWeight(serviceName, blueVersion, blueWeight); err != nil {
		return fmt.Errorf("failed to update blue version weight: %w", err)
	}

	return nil
}

func (p *ProxyManager) RemoveVersion(serviceName, version string) error {
	containerName := fmt.Sprintf("%s-%s", serviceName, version)
	cmd := NewRemoteCommand(p.ssh, fmt.Sprintf("podman rm -f %s", containerName))
	return cmd.Run()
}

func (p *ProxyManager) buildLabelsArg(labels map[string]string) string {
	var args []string
	for k, v := range labels {
		args = append(args, fmt.Sprintf("--label '%s=%s'", k, v))
	}
	return strings.Join(args, " ")
}

func (p *ProxyManager) buildEnvArg(env map[string]string) string {
	var args []string
	for k, v := range env {
		args = append(args, fmt.Sprintf("--env '%s=%s'", k, v))
	}
	return strings.Join(args, " ")
}

func (p *ProxyManager) buildNetworksArg(networks []string) string {
	var args []string
	for _, network := range networks {
		args = append(args, fmt.Sprintf("--network %s", network))
	}
	return strings.Join(args, " ")
}

func (p *ProxyManager) updateWeight(serviceName, version string, weight int) error {
	containerName := fmt.Sprintf("%s-%s", serviceName, version)
	weightLabel := fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.weight=%d", serviceName, weight)
	
	cmd := NewRemoteCommand(p.ssh, fmt.Sprintf(
		"podman update --label-add '%s' %s",
		weightLabel,
		containerName,
	))
	
	return cmd.Run()
}

func (p *ProxyManager) RemoveService(serviceName string) error {
	cmd := NewRemoteCommand(p.ssh, fmt.Sprintf(
		"%s remove %s",
		p.proxyBin,
		serviceName,
	))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove service %s: %w", serviceName, err)
	}

	return nil
}

func (p *ProxyManager) waitForHealthyContainer(containerName string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	interval := 100 * time.Millisecond // Poll more frequently in tests

	for time.Now().Before(deadline) {
		cmd := NewRemoteCommand(p.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Health.Status}}' %s",
			containerName,
		))
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "healthy" {
			return nil
		}
		time.Sleep(interval)
	}
	return fmt.Errorf("container %s did not become healthy within timeout", containerName)
}

func (p *ProxyManager) rollbackWeights(serviceName, blueVersion, greenVersion string, blueWeight, greenWeight int) error {
	if err := p.updateWeight(serviceName, blueVersion, blueWeight); err != nil {
		return err
	}
	if err := p.updateWeight(serviceName, greenVersion, greenWeight); err != nil {
		return err
	}
	return nil
}

func (p *ProxyManager) validateService(service ProxyService) error {
	if service.Name == "" {
		return fmt.Errorf("service name is required")
	}
	if service.Version == "" {
		return fmt.Errorf("version is required")
	}
	if service.Port != "" {
		if _, err := strconv.Atoi(service.Port); err != nil {
			return fmt.Errorf("invalid port: %s", service.Port)
		}
	}
	if service.Weight < 0 || service.Weight > 100 {
		return fmt.Errorf("weight must be between 0 and 100")
	}
	for _, domain := range service.Domains {
		if strings.Contains(domain, " ") {
			return fmt.Errorf("invalid domain: %s", domain)
		}
	}
	return nil
}

func (p *ProxyManager) waitForHealthy(containerName string) error {
	deadline := time.Now().Add(p.healthCheckTimeout)
	for time.Now().Before(deadline) {
		cmd := NewRemoteCommand(p.ssh, fmt.Sprintf(
			"podman inspect --format '{{.State.Health.Status}}' %s",
			containerName,
		))
		output, err := cmd.Output()
		if err == nil && strings.TrimSpace(string(output)) == "healthy" {
			return nil
		}
		select {
		case <-p.ctx.Done():
			return fmt.Errorf("context cancelled while waiting for container to be healthy")
		case <-time.After(p.healthCheckInterval):
			continue
		}
	}
	return fmt.Errorf("container failed to become healthy within timeout")
}
