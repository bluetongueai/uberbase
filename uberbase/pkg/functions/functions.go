package functions

import (
	"log"
	"os"

	"github.com/bluetongueai/uberbase/uberbase/pkg/logging"
)

type FunctionsConfig struct {
	Build string
	Pull  []string
}

var fClient client

func Init(config FunctionsConfig) error {
	log.SetOutput(os.Stdout)

	var err error
	fClient, err = newClient()
	if err != nil {
		logging.Logger.Error("Could not get container client")
	}

	// build all images configured
	subdirs, err := os.ReadDir(config.Build)
	if err != nil {
		logging.Logger.Warnf("Could not read build directory: %v", err)
	} else {
		for _, imageContext := range subdirs {
			imageName := imageContext.Name()
			containerfile := config.Build + "/" + imageContext.Name() + "/Dockerfile"
			err := fClient.Build("uberbase/"+imageName, containerfile, config.Build+"/"+imageContext.Name())
			if err != nil {
				logging.Logger.Warnf("Failed to build image %s, %s, %s", imageName, containerfile, config.Build+"/"+imageContext.Name())
			}
		}
	}

	// login to container registry if DOCKER_TOKEN set on environment
	if os.Getenv("DOCKER_TOKEN") != "" && os.Getenv("DOCKER_USER") != "" {
		logging.Logger.Info("Logging into container registry")
		stdOut, stdErr, err := fClient.container("login", "docker.io", "-u", os.Getenv("DOCKER_USER"), "-p", os.Getenv("DOCKER_TOKEN"))
		if err != nil {
			logging.Logger.Errorf("Failed to login to container registry: %v\n%s\n%s", err, stdOut, stdErr)
		}
		logging.Logger.Infof("Logged into container registry: %s", stdOut)
	}

	for _, image := range config.Pull {
		logging.Logger.Infof("Pulling image %s", image)
		stdOut, stdErr, err := fClient.Pull(image, true)
		if err != nil {
			logging.Logger.Errorf("Failed to pull image: %v\n%s\n%s", err, stdOut, stdErr)
		}
		logging.Logger.Infof("Pulled image %s: %s", image, stdOut)
	}

	return nil
}

func Run(imageName string, detatch bool, env map[string]string, params ...string) (string, string, error) {
	logging.Logger.Infof("Running container %s", imageName)
	return fClient.Run(imageName, detatch, env, params...)
}

func Stop(containerId string) (string, string, error) {
	logging.Logger.Infof("Stopping container %s", containerId)
	return fClient.Stop(containerId)
}
