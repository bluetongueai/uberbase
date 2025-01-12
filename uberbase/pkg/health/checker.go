package health

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/bluetongueai/uberbase/uberbase/pkg/containers"
	"github.com/bluetongueai/uberbase/uberbase/pkg/traefik"
)

// Add HTTP health check configuration
type HTTPHealthCheck struct {
	URL      string
	Method   string
	Path     string
	Status   int
	Interval time.Duration
	Timeout  time.Duration
	Headers  map[string]string
}

type HealthChecker struct {
	containerMgr *containers.ContainerManager
}

func NewHealthChecker(containerMgr *containers.ContainerManager) *HealthChecker {
	return &HealthChecker{
		containerMgr: containerMgr,
	}
}

func (h *HealthChecker) WaitForContainers(ctx context.Context, services map[string]containers.ComposeServiceOverride) (chan bool, error) {
	checks := make([]func() bool, 0, len(services))

	for _, service := range services {
		svc := service // Create a new variable to avoid closure issues
		checks = append(checks, func() bool {
			info, err := h.containerMgr.Inspect(svc.Name)
			return err == nil && info.State.Status == "running"
		})
	}

	return h.waitForAll(checks...)
}

func (h *HealthChecker) WaitForHTTPHealthChecks(ctx context.Context, services map[string]traefik.TraefikService) (chan bool, error) {
	checks := make([]func() bool, 0)

	for _, service := range services {
		if service.LoadBalancer != nil && service.LoadBalancer.HealthCheck.Path != "" {
			for _, server := range service.LoadBalancer.Servers {
				check := h.createHTTPHealthCheck(HTTPHealthCheck{
					URL:      server.URL,
					Method:   service.LoadBalancer.HealthCheck.Method,
					Path:     service.LoadBalancer.HealthCheck.Path,
					Status:   service.LoadBalancer.HealthCheck.Status,
					Headers:  service.LoadBalancer.HealthCheck.Headers,
					Interval: 1 * time.Second, // Default interval
					Timeout:  5 * time.Second, // Default timeout
				})
				checks = append(checks, check)
			}
		}
	}

	if len(checks) == 0 {
		// If no health checks defined, return immediately successful
		ch := make(chan bool, 1)
		ch <- true
		close(ch)
		return ch, nil
	}

	return h.waitForAll(checks...)
}

func (h *HealthChecker) waitForAll(fs ...func() bool) (chan bool, error) {
	healthy := make(chan bool)

	go func() {
		for {
			var wg sync.WaitGroup
			results := make(chan bool, len(fs))

			// Run all checks in parallel
			for _, f := range fs {
				wg.Add(1)
				go func(check func() bool) {
					defer wg.Done()
					results <- check()
				}(f)
			}

			// Close results after all checks complete
			go func() {
				wg.Wait()
				close(results)
			}()

			// Check if all results are true
			allHealthy := true
			for result := range results {
				if !result {
					allHealthy = false
					break
				}
			}

			if allHealthy {
				healthy <- true
				close(healthy)
				return
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return healthy, nil
}

func (h *HealthChecker) createHTTPHealthCheck(config HTTPHealthCheck) func() bool {
	client := &http.Client{
		Timeout: config.Timeout,
	}

	return func() bool {
		req, err := http.NewRequest(config.Method, fmt.Sprintf("%s%s", config.URL, config.Path), nil)
		if err != nil {
			return false
		}

		// Add headers if specified
		for key, value := range config.Headers {
			req.Header.Add(key, value)
		}

		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()

		// If no specific status is required, accept any 2xx status
		if config.Status == 0 {
			return resp.StatusCode >= 200 && resp.StatusCode < 300
		}

		return resp.StatusCode == config.Status
	}
}
