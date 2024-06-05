package functions

import (
	"log"
	"os"
	"time"

	images "github.com/containerd/containerd/v2/core/images"
)

type containerPoolConfig struct {
	InitialSize int
	MaxSize     int
	Images      []string
}

type containerPoolItem struct {
	Container *container
	ImageName string
}

type containerPool struct {
	Containers map[string]containerPoolItem
	Images     map[string]*images.Image
	Config     containerPoolConfig
	Client     client
	available  map[string]bool
	running    map[string]string
}

func newContainerPool(poolConfig containerPoolConfig) (containerPool, error) {
	log.Println("initializing containerd client")
	client, err := newClient()
	if err != nil {
		panic("failed to initialize containerd client")
	}

	log.Printf("pre-building images")
	buildImages(client, builderConfig{
		ImageDirPath: os.Getenv("UBERBASE_IMAGE_PATH"),
	})

	pool := containerPool{
		Containers: make(map[string]containerPoolItem),
		Images:     make(map[string]*images.Image),
		Config:     poolConfig,
		Client:     client,
	}

	log.Println("loading images")
	err = pool.loadImages()
	return pool, nil
}

func (p containerPool) Run(imageName string, params ...string) (string, error) {
	container, err := p.getNextContainer(imageName)
	if err != nil {
		return "", err
	}

	var output string
	if len(params) == 0 {
		output, err = (*container).Run()
	} else {
		output, err = (*container).Exec(params...)
	}
	if err != nil {
		return "", err
	}

	go p.reapPool()

	return output, nil
}

func (p containerPool) Shutdown() {
	log.Println("shutting down container pool")
	for _, containerItem := range p.Containers {
		(*containerItem.Container).Stop()
		(*containerItem.Container).Remove()
	}
	log.Printf("container pool shutdown")
}

func (p containerPool) loadImages() error {
	log.Printf("pulling images %v\n", p.Config.Images)
	for i := 0; i < len(p.Config.Images); i++ {
		imageName := p.Config.Images[i]
		if p.Images[imageName] == nil {
			err := p.pullImage(imageName)
			if err != nil {
				return err
			}
			log.Printf("successfully pulled and registered image %s\n", imageName)
		}
	}

	for i := 0; i < len(p.Config.Images); i++ {
		imageName := p.Config.Images[i]
		log.Printf("initializing %d containers for image %s\n", p.Config.InitialSize, imageName)
		for j := 0; j < p.Config.InitialSize; j++ {
			log.Printf("initializing container %d for image %s\n", j, imageName)
			container, err := newContainer(p.Client, imageName)
			if err != nil {
				return err
			}
			log.Printf("registering container %s for image %s\n", container.ID(), imageName)
			p.Containers[container.ID()] = containerPoolItem{
				Container: &container,
				ImageName: imageName,
			}
			log.Printf("container %s for image %s ready\n", container.ID(), imageName)
		}
	}

	log.Printf("loaded %d images with pool size %d\n", len(p.Config.Images), len(p.Config.Images)*p.Config.InitialSize)

	return nil
}

func (p containerPool) pullImage(image_url string) error {
	log.Printf("pulling image %s\n", image_url)
	log.Printf("client: %v\n", p.Client)
	err := p.Client.Pull(image_url)
	if err != nil {
		return err
	}
	log.Printf("successfully pulled %s image\n", image_url)
	return nil
}

func (p containerPool) getNextContainer(imageName string) (*container, error) {
	log.Printf("finding next available container for image %s\n", imageName)

	image_url := "docker.io/bluetongueai/functions-" + imageName + ":latest"

	for _, containerItem := range p.Containers {
		if containerItem.ImageName != image_url {
			continue
		}

		log.Printf("found container %s", containerItem.Container.ID())
		return containerItem.Container, nil
	}

	// if no containers are available, create a new one and add it to the pool
	// then return it
	log.Printf("no containers available for image %s, creating new container\n", imageName)
	container, err := newContainer(p.Client, imageName)
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return nil, err
	}

	log.Printf("registering new container %s for image %s\n", container.ID(), imageName)
	p.Containers[container.ID()] = containerPoolItem{
		Container: &container,
		ImageName: imageName,
	}

	log.Printf("found container %s", container.ID())
	return &container, nil
}

func (p containerPool) reapPool() error {
	// iterate over all containers in the pool
	// if the container has been running for too long, stop it
	// if the pool is larger than the initial size, delete the container
	counts := make(map[string]int)
	reapCandidates := make(map[string]*container)

	log.Printf("reaping pool\n")
	for _, containerItem := range p.Containers {
		counts[containerItem.ImageName] += 1
		// if a container has been running for more than 24 hours, stop it
		if containerItem.Container.Status == containerStatusRunning {
			wall := containerItem.Container.StartedAt.Add(time.Hour * 24)
			if time.Now().After(wall) {
				log.Printf("container %s has been running for more than 24 hours, stopping\n", containerItem.Container.ID())
				(*containerItem.Container).Stop()
				log.Printf("container %s reset, status %d\n", containerItem.Container.ID(), containerItem.Container.Status)
			}
		}

		// if the pool is larger than the initial size, add the container
		// to the reapCandidates map
		if counts[containerItem.ImageName] > p.Config.MaxSize {
			log.Printf("container %s is a candidate for reaping\n", containerItem.Container.ID())
			reapCandidates[containerItem.Container.ID()] = containerItem.Container
		}
	}

	// reap the candiates
	log.Printf("reaping %d containers\n", len(reapCandidates))
	for _, container := range reapCandidates {
		err := p.reapContainer(container)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p containerPool) reapContainer(container *container) error {
	log.Printf("reaping container %s\n", (*container).ID())
	// stop the container
	(*container).Stop()
	// delete the container
	(*container).Remove()
	// remove the container from the available pool
	delete(p.available, (*container).ID())
	// remove the container from the running pool
	delete(p.running, (*container).ID())
	// remove the container from the containers pool
	delete(p.Containers, (*container).ID())
	log.Printf("container %s reaped\n", (*container).ID())
	return nil
}

func (p containerPool) getContainersByImageName(imageName string) ([]*container, error) {
	var containers []*container
	for _, containerItem := range p.Containers {
		if containerItem.ImageName == imageName {
			containers = append(containers, containerItem.Container)
		}
	}
	return containers, nil
}
