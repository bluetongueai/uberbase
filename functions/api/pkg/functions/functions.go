package functions

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type FunctionsConfig struct {
	Build  string
	Images []string
}

var fClient client
var initialized bool

func Init(config FunctionsConfig) error {
	log.SetOutput(os.Stdout)

	initLima()
	initialized = true

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
	log.Printf("booting compose stack")
	stdOut, stdErr, err := fClient.nerdctl("compose", "up", "-d")
	if err != nil {
		log.Fatalf("failed to start compose stack: %v\n%s\n%s", err, stdOut, stdErr)
	}
	log.Printf("compose stack started: %s", stdOut)

	// capture sigs
	log.Printf("hooking into OS signals to gracefully shutdown")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		Shutdown()
		os.Exit(0)
	}()

	log.Println("container pool initialized")
	return nil
}

func Shutdown() {
	fClient.nerdctl("compose", "down")
}

func Run(imageName string, params ...string) (string, error) {
	if !initialized {
		return "", errors.New("functions not initialized")
	}

	stdOut, stdErr, err := fClient.Run(imageName, params...)
	if err != nil {
		return "", err
	}

	if stdErr != "" {
		return stdErr, nil
	}

	return stdOut, nil
}
