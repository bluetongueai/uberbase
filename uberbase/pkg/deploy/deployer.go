package deploy

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/bluetongueai/uberbase/uberbase/pkg/containers"
	"github.com/bluetongueai/uberbase/uberbase/pkg/core"
	"github.com/bluetongueai/uberbase/uberbase/pkg/environment"
	git "github.com/bluetongueai/uberbase/uberbase/pkg/git"
	"github.com/bluetongueai/uberbase/uberbase/pkg/health"
	"github.com/bluetongueai/uberbase/uberbase/pkg/loadbalancer"
	"github.com/bluetongueai/uberbase/uberbase/pkg/logging"
	"github.com/bluetongueai/uberbase/uberbase/pkg/state"
)

// Deployer orchestrates the deployment process
type Deployer struct {
	compose            *containers.ComposeProject
	gitManager         *git.GitManager
	localContainerMgr  *containers.ContainerManager
	remoteContainerMgr *containers.ContainerManager
	stateManager       *state.StateManager
	healthChecker      *health.HealthChecker
	trafficManager     *loadbalancer.TrafficManager
	environmentManager *environment.EnvironmentManager
	localExecutor      core.Executor
	remoteExecutor     core.Executor
	localWorkDir       string
	remoteWorkDir      string
}

func NewDeployer(localExecutor core.Executor, remoteExecutor core.Executor, compose *containers.ComposeProject, localWorkDir, remoteWorkDir string) (*Deployer, error) {
	logging.Logger.Debug("Verifying local deployment environment requirements")
	if err := localExecutor.Verify(); err != nil {
		return nil, err
	}
	logging.Logger.Debug("Local environment verification complete")

	logging.Logger.Debug("Verifying remote deployment environment requirements")
	if !remoteExecutor.Test() {
		return nil, fmt.Errorf("could not connect to remote server")
	}
	if err := remoteExecutor.Verify(); err != nil {
		return nil, err
	}
	logging.Logger.Debug("Remote environment verification complete")
	localContainerMgr, err := containers.NewContainerManager(localExecutor, compose)
	if err != nil {
		return nil, fmt.Errorf("failed to create local container manager: %w", err)
	}
	remoteContainerMgr, err := containers.NewContainerManager(remoteExecutor, compose)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote container manager: %w", err)
	}
	stateManager := state.NewStateManager(remoteWorkDir, remoteExecutor)
	gitManager, err := git.NewGitManager(localExecutor.(*core.LocalExecutor), remoteExecutor.(*core.RemoteExecutor), localWorkDir, remoteWorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create git manager: %w", err)
	}
	healthChecker := health.NewHealthChecker(remoteContainerMgr)
	trafficManager, err := loadbalancer.NewTrafficManager(remoteContainerMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic manager: %w", err)
	}

	environmentManager := environment.NewEnvironmentManager(remoteExecutor.(*core.RemoteExecutor))

	logging.Logger.Debug("Deployment components initialized successfully")
	return &Deployer{
		compose:            compose,
		gitManager:         gitManager,
		localContainerMgr:  localContainerMgr,
		remoteContainerMgr: remoteContainerMgr,
		stateManager:       stateManager,
		healthChecker:      healthChecker,
		trafficManager:     trafficManager,
		environmentManager: environmentManager,
		localExecutor:      localExecutor,
		remoteExecutor:     remoteExecutor,
		localWorkDir:       localWorkDir,
		remoteWorkDir:      remoteWorkDir,
	}, nil
}

func (d *Deployer) DeployProject() (err error) {
	ctx := context.Background()
	rm := NewRollbackManager(d.stateManager)

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("deployment panic: %v", r)
			if rollbackErr := rm.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("%v, rollback also failed: %v", err, rollbackErr)
			}
		}
	}()

	newVersion, err := d.gitManager.GetCurrentCommit()
	if err != nil {
		return fmt.Errorf("failed to get current git commit: %w", err)
	}
	containerTag := containers.ContainerTag(string(newVersion))

	services := []string{}
	for _, service := range d.compose.Project.Services {
		services = append(services, service.Name)
	}
	logging.LogKeyValues("Building and pushing containers", [][2]string{
		{"tag", string(containerTag)},
		{"services", strings.Join(services, ", ")},
	})

	if _, err := d.localContainerMgr.Build(containerTag); err != nil {
		return fmt.Errorf("failed to build new versions: %w", err)
	}

	if _, err := d.localContainerMgr.Push(containerTag); err != nil {
		return fmt.Errorf("failed to push new versions: %w", err)
	}

	if _, err := d.remoteExecutor.Exec("mkdir -p " + d.remoteWorkDir); err != nil {
		return fmt.Errorf("failed to create remote work directory: %w", err)
	}

	currentState, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load current state: %w", err)
	}

	// TODO: sync secrets from the environment into the remote environment
	if err := d.environmentManager.DeployEnv(containerTag); err != nil {
		return fmt.Errorf("failed to deploy environment: %w", err)
	}

	rm.AddRollbackStep(
		"rollback-env",
		func(ctx context.Context) error {
			return d.environmentManager.DeployEnv(currentState.Tag)
		},
		func(ctx context.Context) error {
			return fmt.Errorf("failed to rollback environment file")
		},
	)

	if err := d.remoteExecutor.SendFile(d.compose.LocalFilePath, filepath.Join(d.remoteWorkDir, "docker-compose.yml")); err != nil {
		return fmt.Errorf("failed to send docker-compose.yml to remote server: %w", err)
	}
	d.compose.RemoteFilePath = filepath.Join(d.remoteWorkDir, "docker-compose.yml")

	// build a dynamic override file
	override := containers.NewComposeOverride(d.compose, containerTag)
	overrideFilePath, err := override.WriteToFile(d.remoteExecutor.(*core.RemoteExecutor), d.remoteWorkDir)
	if err != nil {
		return fmt.Errorf("failed to write override file: %w", err)
	}

	rm.AddRollbackStep(
		"rollback-override",
		func(ctx context.Context) error {
			if _, err := d.remoteExecutor.Exec("rm -f " + overrideFilePath); err != nil {
				return fmt.Errorf("failed to remove override file: %w", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			if _, err := d.remoteExecutor.Exec("test -f " + overrideFilePath); err != nil {
				return nil
			}
			return fmt.Errorf("override file still exists after rollback")
		},
	)

	// pull the new containers
	logging.Logger.Info("Pulling new containers")
	if _, err := d.remoteContainerMgr.Pull(containerTag); err != nil {
		return fmt.Errorf("failed to pull new containers: %w", err)
	}

	// bring up the new containers
	overrideServices := []string{}
	for _, service := range override.Services {
		overrideServices = append(overrideServices, service.Name)
	}
	logging.Logger.Infof("Starting new containers: %s", strings.Join(overrideServices, ", "))

	_, err = d.remoteContainerMgr.Up(overrideFilePath)
	if err != nil {
		return fmt.Errorf("failed to bring up new containers: %w", err)
	}

	logging.Logger.Debug("Waiting for container health checks",
		"services", override.Services,
		"timeout", "10s")

	healthy, err := d.healthChecker.WaitForContainers(context.Background(), override.Services)
	if err != nil {
		return fmt.Errorf("failed to wait for new containers to be healthy: %w", err)
	}
	select {
	case <-healthy:
		logging.Logger.Info("New containers healthy")
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timed out waiting for new containers to be healthy")
	}

	rm.AddRollbackStep(
		"rollback-up",
		func(ctx context.Context) error {
			failedServices := []string{}
			for _, service := range override.Services {
				failedServices = append(failedServices, service.Name)
			}
			if _, err := d.remoteContainerMgr.Down(failedServices, overrideFilePath); err != nil {
				return fmt.Errorf("failed to bring down new containers: %w", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			for _, service := range override.Services {
				if info, err := d.remoteContainerMgr.Inspect(service.Name); err == nil && info.State.Status == "running" {
					return fmt.Errorf("container %s still running after rollback", service.Name)
				}
			}
			return nil
		},
	)

	logging.Logger.Info("Updating traffic routing")
	if err := d.trafficManager.Load(); err != nil {
		return fmt.Errorf("failed to load traffic manager: %w", err)
	}
	if err := d.trafficManager.Deploy(context.Background(), &currentState, containerTag); err != nil {
		return fmt.Errorf("failed to route traffic: %w", err)
	}

	rm.AddRollbackStep(
		"traffic-rollback",
		func(ctx context.Context) error {
			if err := d.trafficManager.Deploy(ctx, &currentState, currentState.Tag); err != nil {
				return fmt.Errorf("failed to rollback traffic: %w", err)
			}
			return nil
		},
		func(ctx context.Context) error {
			// Verify traffic routing
			if currentState.Tag == "" {
				// We've never routed traffic before, so we don't have anything to rollback
				return nil
			}
			for _, config := range currentState.Traefik.Configs {
				healthy, err := d.healthChecker.WaitForHTTPHealthChecks(ctx, config.HTTP.Services)
				if err != nil {
					return fmt.Errorf("health check failed after traffic rollback: %w", err)
				}
				select {
				case <-healthy:
					return nil
				case <-time.After(30 * time.Second):
					return fmt.Errorf("timeout waiting for health checks after traffic rollback")
				}
			}
			return nil
		},
	)

	// bring down the old containers
	oldContainers := []string{}
	for _, service := range currentState.Compose.Services {
		if service.ContainerName == fmt.Sprintf("%s-%s", service.ServiceName, string(containerTag)) {
			continue
		}
		oldContainers = append(oldContainers, service.ContainerName)
	}
	if _, err := d.remoteContainerMgr.Down(oldContainers, overrideFilePath); err != nil {
		return fmt.Errorf("failed to bring down old containers, environment may be inconsistent: %w, %v", err, oldContainers)
	}

	logging.Logger.Info("Cleaning up old containers", "count", fmt.Sprintf("%d", len(oldContainers)), "containers", strings.Join(oldContainers, ", "))

	if _, err := d.remoteContainerMgr.Down(oldContainers, overrideFilePath); err != nil {
		return fmt.Errorf("failed to bring down old containers, environment may be inconsistent: %w, %v", err, oldContainers)
	}

	logging.Logger.Info("Cleaning up old environment")
	if err := d.environmentManager.CommitEnv(currentState.Tag); err != nil {
		return fmt.Errorf("failed to commit environment: %w", err)
	}

	logging.Logger.Info("Updating deployment state")

	if err := d.stateManager.Update(override.Services, d.trafficManager.GetDynamicConfigs(), containerTag); err != nil {
		return fmt.Errorf("failed to create new state: %w", err)
	}
	if err := d.stateManager.Save(); err != nil {
		return fmt.Errorf("failed to save new state, this environment cannot be a rollback target: %w", err)
	}

	logging.Logger.Info("Deployment completed successfully", "version", containerTag)
	return nil
}
