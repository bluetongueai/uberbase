package functions

import (
	"log"
	"os"
)

type FunctionsConfig struct {
	Build  string
	Images []string
}

var fClient client

func Init(config FunctionsConfig) error {
	log.SetOutput(os.Stdout)

	var err error
	fClient, err = newClient()
	if err != nil {
		log.Fatalf("could not get lima/containerd client")
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

	for _, image := range config.Images {
		fClient.Pull(image, false)
		if err != nil {
			log.Fatalf("failed to pull image %s: %v", image, err)
		}
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

func Run(imageName string, params ...string) (string, error) {
	stdOut, stdErr, err := fClient.Run(imageName, params...)
	if err != nil {
		return "", err
	}

	if stdErr != "" {
		return stdErr, nil
	}

	return stdOut, nil
}
