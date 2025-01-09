package deploy

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/containers"
	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	git "github.com/bluetongueai/uberbase/deploy/pkg/git"
	"github.com/bluetongueai/uberbase/deploy/pkg/health"
	"github.com/bluetongueai/uberbase/deploy/pkg/loadbalancer"
	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	"github.com/bluetongueai/uberbase/deploy/pkg/state"
)

// Deployer orchestrates the deployment process
type Deployer struct {
	compose            *containers.ComposeProject
	gitManager         *git.GitManager
	localContainerMgr  *containers.ContainerManager
	remoteContainerMgr *containers.ContainerManager
	stateManager       *state.StateManager
	backupManager      *BackupManager
	healthChecker      *health.HealthChecker
	trafficManager     *loadbalancer.TrafficManager
	localExecutor      core.Executor
	remoteExecutor     core.Executor
}

func NewDeployer(localExecutor core.Executor, remoteExecutor core.Executor, compose *containers.ComposeProject, localWorkDir, remoteWorkDir string) (*Deployer, error) {
	logging.Logger.Info("Checking local environment for requirements...")
	if err := localExecutor.Verify(); err != nil {
		return nil, err
	}
	logging.Logger.Info("Local is ok")

	logging.Logger.Info("Checking remote environment for requirements...")
	if err := remoteExecutor.Verify(); err != nil {
		return nil, err
	}
	logging.Logger.Info("Remote is ok")

	logging.Logger.Debugf("Creating local container manager")
	localContainerMgr, err := containers.NewContainerManager(localExecutor, compose, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create local container manager: %w", err)
	}

	logging.Logger.Debugf("Creating remote container manager")
	remoteContainerMgr, err := containers.NewContainerManager(remoteExecutor, compose, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote container manager: %w", err)
	}

	logging.Logger.Debugf("Creating state manager")
	stateManager := state.NewStateManager(remoteWorkDir, remoteExecutor)
	if _, err := stateManager.Load(); err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	gitManager, err := git.NewGitManager(localExecutor.(*core.LocalExecutor), remoteExecutor.(*core.RemoteExecutor), localWorkDir, remoteWorkDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create git manager: %w", err)
	}

	healthChecker := health.NewHealthChecker(remoteContainerMgr)

	trafficManager, err := loadbalancer.NewTrafficManager(remoteContainerMgr)
	if err != nil {
		return nil, fmt.Errorf("failed to create traffic manager: %w", err)
	}

	backupManager := NewBackupManager(remoteContainerMgr)

	return &Deployer{
		compose:            compose,
		gitManager:         gitManager,
		localContainerMgr:  localContainerMgr,
		remoteContainerMgr: remoteContainerMgr,
		stateManager:       stateManager,
		backupManager:      backupManager,
		healthChecker:      healthChecker,
		trafficManager:     trafficManager,
		localExecutor:      localExecutor,
		remoteExecutor:     remoteExecutor,
	}, nil
}

func (d *Deployer) DeployProject() error {
	// get version to deploy
	newVersion, err := d.gitManager.GetCurrentCommit()
	if err != nil {
		return fmt.Errorf("failed to get current git commit: %w", err)
	}
	containerTag := containers.ContainerTag(string(newVersion))

	logging.Logger.Infof("Building, tagging and pushing version %s", containerTag)

	// build the new versions of all containers through local container manager
	if _, err := d.localContainerMgr.Build(containerTag); err != nil {
		return fmt.Errorf("failed to build new versions: %w", err)
	}

	// push the new versions to the registry with local container manager
	if _, err := d.localContainerMgr.Push(containerTag); err != nil {
		return fmt.Errorf("failed to push new versions: %w", err)
	}

	logging.Logger.Info("Preparing remote environment")
	// on the remote, clone the repo if it doesn't exist
	if err := d.gitManager.Fetch(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}

	// on the remote, pull the new versions
	if _, err := d.remoteContainerMgr.Pull(containerTag); err != nil {
		return fmt.Errorf("failed to pull new versions: %w", err)
	}

	// load the current state
	currentState, err := d.stateManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load current state: %w", err)
	}

	// build a dynamic override file
	override := containers.NewComposeOverride(d.compose, containerTag)
	overrideFilePath, err := override.WriteToFile(d.compose.FilePath + ".override")
	if err != nil {
		return fmt.Errorf("failed to write override file: %w", err)
	}

	// bring up the new containers
	logging.Logger.Info("Bringing up new containers")
	_, err = d.remoteContainerMgr.Up(overrideFilePath)
	if err != nil {
		return fmt.Errorf("failed to bring up new containers: %w", err)
	}

	logging.Logger.Info("Waiting for new containers to be healthy")
	// wait for the new tags to be healthy
	healthy, err := d.healthChecker.WaitForContainers(context.Background(), override.Services)
	if err != nil {
		return fmt.Errorf("failed to wait for new containers to be healthy: %w", err)
	}
	select {
	case <-healthy:
		break
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timed out waiting for new containers to be healthy")
	}

	logging.Logger.Info("Routing traffic to new containers")
	if err := d.trafficManager.Deploy(context.Background(), &currentState, containerTag); err != nil {
		return fmt.Errorf("failed to route traffic: %w", err)
	}

	// bring down the old containers
	oldContainers := []string{}
	for _, service := range currentState.Compose.Services {
		if service.ContainerName == fmt.Sprintf("%s-%s", service.ServiceName, string(containerTag)) {
			continue
		}
		oldContainers = append(oldContainers, service.ContainerName)
	}
	if _, err := d.remoteContainerMgr.Down(oldContainers); err != nil {
		return fmt.Errorf("failed to bring down old containers: %w", err)
	}

	// create a new state
	if err := d.stateManager.Update(override.Services, d.trafficManager.GetDynamicConfigs(), containerTag); err != nil {
		return fmt.Errorf("failed to create new state: %w", err)
	}
	if err := d.stateManager.Save(); err != nil {
		return fmt.Errorf("failed to save new state: %w", err)
	}

	// remove the docker override
	if err := os.Remove(overrideFilePath); err != nil {
		return fmt.Errorf("failed to remove docker override file: %w", err)
	}

	return nil
}
