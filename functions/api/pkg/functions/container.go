package functions

import (
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
	ExitedAt      time.Time
	Status        containerStatus
	c             client
}

func newContainer(client client, imageName string) (container, error) {
	c := container{
		ImageName: imageName,
		Status:    containerStatusIdle,
		c:         client,
	}
	containerName, err := client.NewContainer(c.ImageName)
	if err != nil {
		return container{}, err
	}

	c.ContainerName = containerName

	return c, nil
}

func (c container) client() client {
	return c.c
}

func (c container) ID() string {
	return c.ContainerName
}

func (c container) Run() (string, error) {
	c.StartedAt = time.Now()
	c.Status = containerStatusRunning
	stdout, stderr, err := c.client().Run(c.ContainerName)
	if err != nil {
		return "", err
	}
	c.ExitedAt = time.Now()
	c.Status = containerStatusStopped
	if stderr != "" {
		return stderr, nil
	}
	return stdout, nil
}

func (c container) Exec(params ...string) (string, error) {
	stdout, stderr, err := c.client().Exec(c.ContainerName, params...)
	if err != nil {
		return stderr, err
	}
	return stdout, nil
}

func (c container) Stop() error {
	err := c.client().Stop(c.ContainerName)
	if err != nil {
		return err
	}
	return nil
}

func (c container) Remove() error {
	err := c.client().Remove(c.ContainerName)
	if err != nil {
		return err
	}
	return nil
}
