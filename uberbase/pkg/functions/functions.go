package functions

import (
	"log"
	"os"
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
		log.Fatalf("could not get container client")
	}

	// build all images configured
	subdirs, err := os.ReadDir(config.Build)
	if err != nil {
		log.Printf("could not read build directory: %v", err)
	} else {

		for _, imageContext := range subdirs {
			imageName := imageContext.Name()
			containerfile := config.Build + "/" + imageContext.Name() + "/Dockerfile"
			err := fClient.Build("uberbase/"+imageName, containerfile, config.Build+"/"+imageContext.Name())
			if err != nil {
				log.Printf("failed to build image %s, %s, %s", imageName, containerfile, config.Build+"/"+imageContext.Name())
			}
		}
	}

	// login to container registry if DOCKER_TOKEN set on environment
	if os.Getenv("DOCKER_TOKEN") != "" && os.Getenv("DOCKER_USER") != "" {
		log.Printf("logging into container registry")
		stdOut, stdErr, err := fClient.container("login", "docker.io", "-u", os.Getenv("DOCKER_USER"), "-p", os.Getenv("DOCKER_TOKEN"))
		if err != nil {
			log.Fatalf("failed to login to container registry: %v\n%s\n%s", err, stdOut, stdErr)
		}
		log.Printf("logged into container registry: %s", stdOut)
	}

	for _, image := range config.Pull {
		log.Printf("pulling image %s", image)
		stdOut, stdErr, err := fClient.Pull(image, true)
		if err != nil {
			log.Fatalf("failed to pull image: %v\n%s\n%s", err, stdOut, stdErr)
		}
		log.Printf("pulled image %s: %s", image, stdOut)
	}

	return nil
}

func Run(imageName string, detatch bool, env map[string]string, params ...string) (string, string, error) {
	return fClient.Run(imageName, detatch, env, params...)
}

func Stop(containerId string) (string, string, error) {
	return fClient.Stop(containerId)
}
