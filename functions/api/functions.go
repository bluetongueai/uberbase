package main

import (
	"context"
	"fmt"
	"log"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
)

/*
*
* Architecture:
*
* The design of this ...whatever it is... is oriented around fast execution of images.
* To achieve this, we'll maintain a pool of empty containers that are ready to run images.
*
* Given that the set of possible images to run on containers is unknown, but configured at
* runtime, we'll need to calculate the set based on a known dir on the filesystem, then
* pull all the images in that set. Once we have the images, we'll need to construct a
* number of containers for each image. The pool size will determine the per image container
* count. Each container in the pool will need to be created and assigned a snapshot.
*
* Requests to run images on the pool will select an available container that matches the
* image, and run a task on it. The task will be started and the output will be returned.
*
* While the intention of this system is to support ephemeral, short running tasks, we'll
* need to support long running tasks. This will require a mechanism to track the state of
* each container and task, and to clean up after tasks that have been running for too long.
*
* If a request to run an image cannot be satisfied by the pool, a set of containers will be
* created and added to the pool. The pool will dynamically expand to accomodate the new
* containers, and will shrink as containers are deleted, but will never go below the initial
* pool size.
*
 */

type ContainerPoolConfig struct {
	InitialSize int
	MaxSize     int
	Images      []string
}

type ContainerPoolItem struct {
	Container *containerd.Container
	ImageName string
	StartedAt time.Time
}

type ContainerPool struct {
	Containers map[string]ContainerPoolItem
	Images     map[string]*containerd.Image
	Tasks      map[string]*containerd.Task
	Config     ContainerPoolConfig
}

// emulate an enum for ContainerStatus
type ContainerStatus int

const (
	ContainerStatusRunning ContainerStatus = iota
	ContainerStatusStopped
)

const socket = "/run/containerd/containerd.sock"

var client *containerd.Client
var containerContext context.Context
var pool *ContainerPool

func InitContainerd(poolConfig ContainerPoolConfig) error {
	err := initContainerdClient()
	if err != nil {
		return err
	}

	pool = &ContainerPool{
		Containers: make(map[string]ContainerPoolItem),
		Images:     make(map[string]*containerd.Image),
		Tasks:      make(map[string]*containerd.Task),
		Config:     poolConfig,
	}

	for i := 0; i < len(poolConfig.Images); i++ {
		imageName := poolConfig.Images[i]
		if pool.Images[imageName] == nil {
			image, err := pullImage(imageName)
			if err != nil {
				return err
			}

			pool.Images[imageName] = image
		}

		for j := 0; j < poolConfig.InitialSize; j++ {
			container, err := createContainer(imageName)
			if err != nil {
				return err
			}
			pool.Containers[(*container).ID()] = ContainerPoolItem{
				Container: container,
				ImageName: imageName,
			}
		}
	}

	return nil
}

func RunImage(imageName string, params ...string) (uint32, time.Time, error) {
	// if this image is new to the system, pull it
	if pool.Images[imageName] == nil {
		image, err := pullImage(imageName)
		if err != nil {
			return 1, time.Now(), err
		}
		pool.Images[imageName] = image
	}

	container, err := getNextContainer(imageName)
	if err != nil {
		return 1, time.Now(), err
	}

	exitStatus, exitAt, err := runImage(container, params...)
	go reapPool()

	return exitStatus, exitAt, err
}

func initContainerdClient() error {
	c, err := containerd.New(socket)
	if err != nil {
		return err
	}
	defer c.Close()
	ctx := namespaces.WithNamespace(context.Background(), "uberbase")
	containerContext = ctx
	client = c
	return nil
}

func pullImage(image_url string) (*containerd.Image, error) {
	image, err := client.Pull(containerContext, image_url, containerd.WithPullUnpack)
	if err != nil {
		return nil, err
	}
	log.Printf("Successfully pulled %s image\n", image.Name())
	return &image, nil
}

func createContainer(imageName string) (*containerd.Container, error) {
	snapshotName := fmt.Sprintf("%s-snapshot", imageName)

	image := *pool.Images[imageName]

	container, err := client.NewContainer(
		containerContext,
		imageName,
		containerd.WithNewSnapshot(snapshotName, image),
		containerd.WithNewSpec(oci.WithImageConfig(image)),
	)
	if err != nil {
		log.Printf("Failed to create container: %v", err)
		return nil, err
	}

	defer container.Delete(containerContext, containerd.WithSnapshotCleanup)
	log.Printf("Successfully created container with ID %s and snapshot with ID %s-snapshot", container.ID(), image)
	return &container, nil
}

func runImage(container *containerd.Container, params ...string) (uint32, time.Time, error) {
	task, err := (*container).NewTask(containerContext, cio.NewCreator(cio.WithStdio))
	if err != nil {
		return 1, time.Now(), err
	}
	defer task.Delete(containerContext)

	pool.Tasks[(*container).ID()] = &task

	exitStatusC, err := task.Wait(containerContext)
	if err != nil {
		return 1, time.Now(), err
	}

	if err := task.Start(containerContext); err != nil {
		return 1, time.Now(), err
	}

	containerItem := pool.Containers[(*container).ID()]
	containerItem.StartedAt = time.Now()

	status := <-exitStatusC
	code, exitedAt, err := status.Result()
	return code, exitedAt, nil
}

func getNextContainer(imageName string) (*containerd.Container, error) {
	for id, containerItem := range pool.Containers {
		if containerItem.ImageName != imageName {
			continue
		}

		// TODO - ensure tasks remove themselves from the pool on exit
		if pool.Tasks[id] != nil {
			continue
		}

		return containerItem.Container, nil
	}

	// if no containers are available, create a new one and add it to the pool
	// then return it
	container, err := createContainer(imageName)
	if err != nil {
		log.Printf("Failed to create container: %v", err)
		return nil, err
	}

	pool.Containers[(*container).ID()] = ContainerPoolItem{
		Container: container,
		ImageName: imageName,
	}

	return container, nil
}

func reapPool() error {
	// iterate over all containers in the pool
	// if the container has been running for too long, stop it
	// if the pool is larger than the initial size, delete the container
	counts := make(map[string]int)

	for id, containerItem := range pool.Containers {
		counts[containerItem.ImageName] += 1
		task := pool.Tasks[id]
		if task != nil {
			status, err := (*task).Status(containerContext)
			if err != nil {
				return err
			}
			if status.Status == containerd.Running {
				wall := containerItem.StartedAt.Add(time.Hour * 24)
				if time.Now().After(wall) {
					// reap
				}
			}
		}
	}

	for imageName, count := range counts {
		if count > pool.Config.MaxSize {
			delta := pool.Config.MaxSize - count
			containers, err := getContainersByImageName(imageName)
		}
	}

	return nil
}

func getContainersByImageName(imageName string) ([]*containerd.Container, error) {

}
