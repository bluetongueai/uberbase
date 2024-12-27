package deploy

import (
	"fmt"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

// TrafficManager handles traffic routing during deployments
type TrafficManager struct {
	ssh *core.SSHConnection
}

func NewTrafficManager(ssh *core.SSHConnection) *TrafficManager {
	return &TrafficManager{
		ssh: ssh,
	}
}

func (t *TrafficManager) UpdateTraffic(service *Service, blueVersion, greenVersion string, blueWeight, greenWeight int) error {
	// Update Traefik labels for blue version
	if err := t.updateContainerLabels(service.Name, blueVersion, map[string]string{
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.weight", service.Name): fmt.Sprintf("%d", blueWeight),
	}); err != nil {
		return fmt.Errorf("failed to update blue version labels: %w", err)
	}

	// Update Traefik labels for green version
	if err := t.updateContainerLabels(service.Name, greenVersion, map[string]string{
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.weight", service.Name): fmt.Sprintf("%d", greenWeight),
	}); err != nil {
		return fmt.Errorf("failed to update green version labels: %w", err)
	}

	return nil
}

func (t *TrafficManager) updateContainerLabels(serviceName, version string, labels map[string]string) error {
	containerName := fmt.Sprintf("%s-%s", serviceName, version)

	for k, v := range labels {
		cmd := fmt.Sprintf("podman container update --label-add '%s=%s' %s", k, v, containerName)
		if _, err := t.ssh.Exec(cmd); err != nil {
			return fmt.Errorf("failed to update container labels: %w", err)
		}
	}

	return nil
}
