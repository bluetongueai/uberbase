package deploy

import (
	"fmt"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

type HealthCheckConfig struct {
	MaxWaitTime   time.Duration `yaml:"max_wait_time"`
	CheckInterval time.Duration `yaml:"check_interval"`
	FailureAction string        `yaml:"failure_action"` // "rollback" or "continue"
}

type HealthChecker struct {
	ssh     *core.SSHConnection
	metrics *DeploymentMetrics
}

func NewHealthChecker(ssh *core.SSHConnection, metrics *DeploymentMetrics) *HealthChecker {
	return &HealthChecker{
		ssh:     ssh,
		metrics: metrics,
	}
}

func (h *HealthChecker) WaitForHealthy(service *Service, version string, config HealthCheckConfig) error {
	startTime := time.Now()
	containerName := fmt.Sprintf("%s-%s", service.Name, version)
	deadline := time.Now().Add(config.MaxWaitTime)

	ticker := time.NewTicker(config.CheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		healthy, err := h.checkContainerHealth(containerName)
		if err != nil {
			return fmt.Errorf("health check failed: %w", err)
		}

		if healthy {
			duration := time.Since(startTime)
			h.metrics.RecordHealthCheck(true, duration)
			return nil
		}

		if time.Now().After(deadline) {
			h.metrics.RecordHealthCheck(false, time.Since(startTime))
			if config.FailureAction == "rollback" {
				return fmt.Errorf("health check timed out after %v", config.MaxWaitTime)
			}
			core.Logger.Warn("Health check timed out but continuing due to configuration")
			return nil
		}
	}

	return nil
}

func (h *HealthChecker) checkContainerHealth(containerName string) (bool, error) {
	cmd := fmt.Sprintf("podman inspect --format '{{.State.Health.Status}}' %s", containerName)
	output, err := h.ssh.Exec(cmd)
	if err != nil {
		return false, fmt.Errorf("failed to check container health: %w", err)
	}

	return string(output) == "healthy", nil
}
