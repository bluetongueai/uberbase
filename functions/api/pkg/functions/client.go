package functions

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/goombaio/namegenerator"
)

var seed = time.Now().UTC().UnixNano()
var nameGenerator = namegenerator.NewNameGenerator(seed)
var nameCounts = make(map[string]int)

type client struct {
	bin              string
	containerdClient *containerd.Client
	isLima           bool
}

func newClient() (client, error) {
	client := client{}
	log.Println("determining correct command")
	// check if `lima` is available on the host
	path, err := exec.LookPath("lima")
	if err == nil {
		client.bin = path
		client.isLima = true
		log.Printf("using `%s nerdctl` as the container runtime", client.bin)
		log.Printf("isLima: %v", client.isLima)
		return client, nil
	}
	path, err = exec.LookPath("nerdctl")
	if err == nil {
		client.bin = path
		log.Printf("using `%s` as the container runtime", client.bin)
		log.Printf("isLima: %v", client.isLima)
		return client, nil
	}

	log.Printf("neither `lima` nor `nerdctl` is available on the host")
	return client, fmt.Errorf("container runtime not found")
}

func (c client) command(args ...string) (string, string, error) {
	ctx := context.Background()
	var cmd *exec.Cmd
	if c.isLima {
		args = append([]string{"nerdctl"}, args...)
		log.Printf("running command %s %v", c.bin, args)
		cmd = exec.CommandContext(ctx, c.bin, args...)
	} else {
		log.Printf("running command %s %v", c.bin, args)
		cmd = exec.CommandContext(ctx, c.bin, args...)
	}
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer
	err := cmd.Run()
	if err != nil {
		return "", "", err
	}
	return stdoutBuffer.String(), stderrBuffer.String(), nil
}

func (c client) Pull(imageName string, force bool) error {
	if force || !c.imageExists(imageName) {
		log.Printf("fetching containerd image %s", imageName)
		_, _, err := c.command("pull", imageName)
		if err != nil {
			log.Printf("failed to pull image %s: %v", imageName, err)
			return err
		}
		log.Printf("successfully pulled image %s", imageName)
		return nil
	}
	return nil
}

func (c client) imageExists(imageName string) bool {
	_, _, err := c.command("inspect", imageName)
	if err != nil {
		return false
	}
	return true
}

func (c client) Build(imageName, dockerfile string, context string) error {
	log.Printf("building containerd image %s", imageName)
	_, _, err := c.command("build", "-t", imageName, "-f", dockerfile, context)
	if err != nil {
		log.Printf("failed to build image %s: %v", imageName, err)
		return err
	}
	log.Printf("successfully built image %s", imageName)
	return nil
}

func (c client) NewContainer(imageName string) (string, error) {
	log.Printf("creating container for image %s", imageName)
	name := nameGenerator.Generate()
	nameCounts[name]++
	name = fmt.Sprintf("%s-%d", name, nameCounts[name])
	_, _, err := c.command("create", "--name", name, imageName)
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return "", err
	}
	log.Printf("successfully created container with ID %s for image %s", name, imageName)
	return name, nil
}

func (c client) Start(containerName string) (string, string, error) {
	log.Printf("running container %s", containerName)
	stdout, stderr, err := c.command("start", "-a", containerName)
	if err != nil {
		log.Printf("failed to start container %s: %v", containerName, err)
		return "", "", err
	}
	log.Printf("successfully ran container %s", containerName)
	return stdout, stderr, nil
}

func (c client) Exec(containerName string, params ...string) (string, string, error) {
	log.Printf("executing command %v in container %s", params, containerName)
	params = append([]string{"exec", containerName}, params...)
	stdout, stderr, err := c.command(params...)
	if err != nil {
		log.Printf("failed to execute command in container %s: %v", containerName, err)
		return "", "", err
	}
	log.Printf("successfully executed command in container %s", containerName)
	return stdout, stderr, nil
}

func (c client) Stop(containerName string) error {
	log.Printf("stopping container %s", containerName)
	_, _, err := c.command("stop", containerName)
	if err != nil {
		log.Printf("failed to stop container %s: %v", containerName, err)
		return err
	}
	log.Printf("successfully stopped container %s", containerName)
	return nil
}

func (c client) Remove(containerName string) error {
	log.Printf("removing container %s", containerName)
	_, _, err := c.command("rm", containerName)
	if err != nil {
		log.Printf("failed to remove container %s: %v", containerName, err)
		return err
	}
	log.Printf("successfully removed container %s", containerName)
	return nil
}

func (c client) Run(imageName string, params ...string) (string, string, error) {
	imageParams := append([]string{"run", imageName}, params...)
	stdout, stderr, err := c.command(imageParams...)
	if err != nil {
		log.Printf("failed to run image %s: %v", imageName, err)
		return "", "", err
	}
	return stdout, stderr, nil
}
