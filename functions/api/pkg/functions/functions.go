package functions

import (
	"log"
	"os"
)

type FunctionsConfig struct {
	Build  string
	Pull []string
}

var fClient client

func Init(config FunctionsConfig) error {
	log.SetOutput(os.Stdout)

	var err error
	fClient, err = newClient()
	if err != nil {
		log.Fatalf("could not get docker client")
	}

	// build all images configured
	subdirs, err := os.ReadDir(config.Build)
	if err != nil {
		log.Printf("could not read build directory: %v", err)
	} else {

		for _, imageContext := range subdirs {
			imageName := imageContext.Name()
			dockerfile := config.Build + "/" + imageContext.Name() + "/Dockerfile"
			err := fClient.Build("uberbase/"+imageName, dockerfile, config.Build+"/"+imageContext.Name())
			if err != nil {
				log.Printf("failed to build image %s, %s, %s", imageName, dockerfile, config.Build+"/"+imageContext.Name())
			}
		}
	}

	// login to docker registry if DOCKER_TOKEN set on environment
	if os.Getenv("DOCKER_TOKEN") != "" && os.Getenv("DOCKER_USER") != "" {
		log.Printf("logging into docker registry")
		stdOut, stdErr, err := fClient.docker("login", "-u", os.Getenv("DOCKER_USER"), "-p", os.Getenv("DOCKER_TOKEN"))
		if err != nil {
			log.Fatalf("failed to login to docker registry: %v\n%s\n%s", err, stdOut, stdErr)
		}
		log.Printf("logged into docker registry: %s", stdOut)
	}

	for _, image := range config.Pull {
		log.Printf("pulling image %s", image)
		stdOut, stdErr, err := fClient.Pull(image, false)
		if err != nil {
			log.Fatalf("failed to pull image: %v\n%s\n%s", err, stdOut, stdErr)
		}
		log.Printf("pulled image %s: %s", image, stdOut)
	}

	// start the docker compose stack
	log.Printf("building compose stack")
	stdOut, stdErr, err := fClient.dockerCompose("build")
	if err != nil {
		log.Fatalf("failed to build compose stack: %v\n%s\n%s", err, stdOut, stdErr)
	}
	log.Printf("compose stack built: %s", stdErr)
	log.Printf("booting compose stack")
	stdOut, stdErr, err = fClient.dockerCompose("up", "-d")
	if err != nil {
		log.Fatalf("failed to start compose stack: %v\n%s\n%s", err, stdOut, stdErr)
	}
	log.Printf("compose stack started: %s", stdErr)

	stdOut, stdErr, err = fClient.dockerCompose("ps")
	log.Printf("compose stack status: %s\n%s", stdOut, stdErr)

	return nil
}

func Shutdown() {
	log.Printf("shutting down compose stack")
	stdout, stderr, err := fClient.dockerCompose("down")
	if err != nil {
		log.Fatalf("failed to shutdown cleanly: %v\n%s\n%s", err, stdout, stderr)
	}
	log.Printf("compose stack shutdown: %s", stdout)
}

func Run(imageName string, params ...string) (string, string, error) {
	return fClient.Run(imageName, params...)
}
