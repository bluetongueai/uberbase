package functions

import (
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type FunctionsConfig struct {
	MinPoolSize int
	MaxPoolSize int
	Images      []string
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

func functionNameToImageUrl(name string) string {
	return "docker.io/bluetongueai/functions-" + name + ":latest"
}
