package loadbalancer

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/containers"
	"github.com/bluetongueai/uberbase/deploy/pkg/health"
	"github.com/bluetongueai/uberbase/deploy/pkg/state"
	"github.com/bluetongueai/uberbase/deploy/pkg/traefik"
)

const (
	healthCheckTimeout = 10 * time.Second
)

// TrafficManager handles the routing and load balancing of traffic between different
// versions of services during deployments.
type TrafficManager struct {
	staticConfig   *traefik.TraefikStaticConfiguration
	dynamicConfigs map[string]*traefik.TraefikDynamicConfiguration
	containerMgr   *containers.ContainerManager
	healthChecker  *health.HealthChecker
}

// NewTrafficManager creates a new TrafficManager instance with the provided container manager.
// It loads the static and dynamic Traefik configurations and initializes the health checker.
func NewTrafficManager(containerMgr *containers.ContainerManager) (*TrafficManager, error) {
	staticConfig, err := traefik.LoadTraefikStaticConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load static config: %w", err)
	}
	dynamicConfigs, err := traefik.LoadTraefikDynamicConfigs()
	if err != nil {
		return nil, fmt.Errorf("failed to load dynamic configs: %w", err)
	}

	healthChecker := health.NewHealthChecker(containerMgr)

	return &TrafficManager{
		staticConfig:   staticConfig,
		dynamicConfigs: dynamicConfigs,
		containerMgr:   containerMgr,
		healthChecker:  healthChecker,
	}, nil
}

// Deploy updates the traffic routing for a new deployment with the given tag.
// It ensures the new services are healthy before updating the routing configuration.
func (t *TrafficManager) Deploy(ctx context.Context, state *state.DeploymentState, tag containers.ContainerTag) error {
	if tag == "" {
		return fmt.Errorf("container tag cannot be empty")
	}

	if state == nil {
		return fmt.Errorf("deployment state cannot be nil")
	}

	if state.Tag == tag {
		return nil
	}

	oldTag := state.Tag

	if t.hasDeployConfigs(tag) {
		return fmt.Errorf("deploy config already exists for tag %s", tag)
	}

	deployConfigs, err := t.createDeployConfigs(tag)
	if err != nil {
		return fmt.Errorf("failed to create deploy config: %w", err)
	}

	for _, config := range deployConfigs {
		healthyChan, err := t.waitForHealthy(ctx, config)
		if err != nil {
			return fmt.Errorf("failed to wait for health: %w", err)
		}

		select {
		case isHealthy := <-healthyChan:
			if !isHealthy {
				return fmt.Errorf("health check failed for service")
			}
		case <-ctx.Done():
			return fmt.Errorf("deployment cancelled: %w", ctx.Err())
		case <-time.After(healthCheckTimeout):
			return fmt.Errorf("timed out waiting for health")
		}
	}

	if err := t.updateRouters(tag); err != nil {
		return fmt.Errorf("failed to update routers: %w", err)
	}

	if err := t.removeDynamicConfigs(oldTag); err != nil {
		return fmt.Errorf("failed to remove dynamic configs: %w", err)
	}

	return nil
}

func (t *TrafficManager) GetDynamicConfigs() map[string]*traefik.TraefikDynamicConfiguration {
	configs := make(map[string]*traefik.TraefikDynamicConfiguration)
	for filename, config := range t.dynamicConfigs {
		if strings.HasSuffix(filename, "-deploy.yml") {
			continue
		}
		configs[filename] = config
	}
	return configs
}

func (t *TrafficManager) hasDeployConfigs(tag containers.ContainerTag) bool {
	for configFile := range t.dynamicConfigs {
		if strings.HasSuffix(configFile, fmt.Sprintf("%s-deploy.yml", string(tag))) {
			return true
		}
	}
	return false
}

func (t *TrafficManager) createDeployConfigs(tag containers.ContainerTag) (map[string]*traefik.TraefikDynamicConfiguration, error) {
	deployConfigs := make(map[string]*traefik.TraefikDynamicConfiguration)
	for configFile, config := range t.dynamicConfigs {
		if strings.HasSuffix(configFile, fmt.Sprintf("%s-deploy.yml", string(tag))) {
			continue
		}
		tagConfig := config.Copy()
		if tagConfig.HTTP == nil || tagConfig.HTTP.Services == nil {
			return nil, fmt.Errorf("invalid configuration: HTTP or Services is nil")
		}
		for _, service := range tagConfig.HTTP.Services {
			for _, server := range service.LoadBalancer.Servers {
				host, port, err := parseURL(server.URL)
				if err != nil {
					return nil, fmt.Errorf("failed to parse URL: %w", err)
				}
				server.URL = fmt.Sprintf("http://%s-%s:%d", host, string(tag), port)
			}
		}
		for _, router := range tagConfig.HTTP.Routers {
			router.Service = fmt.Sprintf("%s-%s", router.Service, string(tag))
		}
		if err := tagConfig.WriteToFile(traefik.DynamicConfigPath, fmt.Sprintf("%s-deploy.yml", string(tag))); err != nil {
			return nil, fmt.Errorf("failed to write tag config: %w", err)
		}
		deployConfigs[fmt.Sprintf("%s.yml", string(tag))] = tagConfig
	}
	return deployConfigs, nil
}

func (t *TrafficManager) updateRouters(tag containers.ContainerTag) error {
	for configFile, config := range t.dynamicConfigs {
		if !strings.HasSuffix(configFile, fmt.Sprintf("%s-deploy.yml", string(tag))) {
			continue
		}
		if config.HTTP != nil && config.HTTP.Routers != nil {
			for _, router := range config.HTTP.Routers {
				router.Service = fmt.Sprintf("%s-%s", router.Service, string(tag))
			}
		}
	}
	return nil
}

func (t *TrafficManager) waitForHealthy(ctx context.Context, config *traefik.TraefikDynamicConfiguration) (<-chan bool, error) {
	healthyChan := make(chan bool, 1)

	go func() {
		defer close(healthyChan)

		httpChan, err := t.healthChecker.WaitForHTTPHealthChecks(ctx, config.HTTP.Services)
		if err != nil {
			healthyChan <- false
			return
		}

		select {
		case result := <-httpChan:
			healthyChan <- result
		case <-ctx.Done():
			healthyChan <- false
		}
	}()

	return healthyChan, nil
}

// parseURL parses a URL string into host and port components.
// The URL should be in the format "host:port".
func parseURL(urlStr string) (string, int, error) {
	urlStr = strings.TrimPrefix(strings.TrimPrefix(urlStr, "http://"), "https://")

	split := strings.Split(urlStr, ":")
	if len(split) != 2 {
		return "", 0, fmt.Errorf("invalid URL format (expected host:port): %s", urlStr)
	}

	port, err := strconv.Atoi(split[1])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port number %q: %w", split[1], err)
	}

	if port < 1 || port > 65535 {
		return "", 0, fmt.Errorf("port number out of range (1-65535): %d", port)
	}

	return split[0], port, nil
}

func (t *TrafficManager) removeDynamicConfigs(tag containers.ContainerTag) error {
	for configFile := range t.dynamicConfigs {
		if strings.HasSuffix(configFile, fmt.Sprintf("%s-deploy.yml", string(tag))) {
			os.Remove(traefik.DynamicConfigPath + configFile)
		}
	}
	return nil
}
