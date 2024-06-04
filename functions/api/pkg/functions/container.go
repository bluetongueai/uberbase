package functions

import (
	"log"
	"time"
)

// go enum for Container status
type containerStatus int

const (
	containerStatusIdle containerStatus = iota
	containerStatusRunning
	containerStatusStopped
)

// Idle when no task is loaded
// Running when task is loaded and running
// Stopped when task is loaded and stopped

type container struct {
	ImageName     string
	ContainerName string
	StartedAt     time.Time
	Status        containerStatus
	c             client
}

func newContainer(client client, imageName string) (container, error) {
	log.Printf("creating containerd container for image %s", imageName)
	c := container{
		ImageName: imageName,
		Status:    containerStatusIdle,
		c:         client,
	}
	containerName, err := client.NewContainer(containerContext, c.ImageName)
	if err != nil {
		return container{}, err
	}

	c.ContainerName = containerName
	log.Printf("successfully created container with ID %s", c.ContainerName)

	return c, nil
}

func (c container) client() client {
	return c.c
}

func (c container) ID() string {
	return c.ContainerName
}

func (c container) Exec(params ...string) (string, error) {
	log.Printf("executing command in container %s with params %v", c.ContainerName, params)
	output, stderr, err := c.client().Exec(c.ContainerName, params...)
	if err != nil {
		log.Printf("failed to execute command in container %s: %v", c.ContainerName, err)
		return stderr, err
	}
	log.Printf("successfully executed command in container %s", c.ContainerName)
	return output, nil
}

func (c container) Stop() error {
	log.Printf("stopping container %s", c.ContainerName)
	err := c.client().Stop(containerContext, c.ContainerName)
	if err != nil {
		log.Printf("failed to stop container %s: %v", c.ContainerName, err)
		return err
	}
	log.Printf("successfully stopped container %s", c.ContainerName)
	return nil
}

func (c container) Remove() error {
	log.Printf("removing container %s", c.ContainerName)
	err := c.client().Remove(containerContext, c.ContainerName)
	if err != nil {
		log.Printf("failed to remove container %s: %v", c.ContainerName, err)
		return err
	}
	log.Printf("successfully removed container %s", c.ContainerName)
	return nil
}
