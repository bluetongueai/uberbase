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
		log.Printf("using `%s` as the container runtime", client.bin)
		return client, nil
	}
	path, err = exec.LookPath("nerdctl")
	if err == nil {
		client.bin = path
		log.Printf("using `%s` as the container runtime", client.bin)
		return client, nil
	}

	log.Printf("neither `lima` nor `nerdctl` is available on the host")
	return client, fmt.Errorf("container runtime not found")
}

func (c client) command(ctx context.Context, args ...string) error {
	log.Printf("running command %s %v", c.bin, args)
	var cmd *exec.Cmd
	if c.isLima {
		args = append([]string{"nerdctl"}, args...)
		cmd = exec.CommandContext(ctx, c.bin, args...)
	} else {
		cmd = exec.CommandContext(ctx, c.bin, args...)
	}
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer
	err := cmd.Run()
	ctx = context.WithValue(ctx, "stdout", stdoutBuffer.String())
	ctx = context.WithValue(ctx, "stderr", stderrBuffer.String())
	if err != nil {
		return err
	}
	return nil
}

func (c client) Pull(ctx context.Context, imageName string) error {
	log.Printf("fetching containerd image %s", imageName)
	err := c.command(ctx, "pull", imageName)
	if err != nil {
		log.Printf("failed to pull image %s: %v", imageName, err)
		return err
	}
	log.Printf("successfully pulled image %s", imageName)
	return nil
}

func (c client) NewContainer(ctx context.Context, imageName string) (string, error) {
	log.Printf("creating container for image %s", imageName)
	name := nameGenerator.Generate()
	err := c.command(ctx, "create", "--name", name, imageName)
	if err != nil {
		log.Printf("failed to create container: %v", err)
		return "", err
	}
	log.Printf("successfully created container with ID %s for image %s", name, imageName)
	return name, nil
}

func (c client) Exec(containerName string, params ...string) (string, string, error) {
	log.Printf("executing command %v in container %s", params, containerName)
	ctx := context.Background()
	params = append([]string{"exec", containerName}, params...)
	err := c.command(ctx, params...)
	if err != nil {
		log.Printf("failed to execute command in container %s: %v", containerName, err)
		return "", "", err
	}
	stdout := ctx.Value("stdout").(string)
	stderr := ctx.Value("stderr").(string)
	log.Printf("successfully executed command in container %s", containerName)
	return stdout, stderr, nil
}

func (c client) Stop(ctx context.Context, containerName string) error {
	log.Printf("stopping container %s", containerName)
	err := c.command(ctx, "stop", containerName)
	if err != nil {
		log.Printf("failed to stop container %s: %v", containerName, err)
		return err
	}
	log.Printf("successfully stopped container %s", containerName)
	return nil
}

func (c client) Remove(ctx context.Context, containerName string) error {
	log.Printf("removing container %s", containerName)
	err := c.command(ctx, "rm", containerName)
	if err != nil {
		log.Printf("failed to remove container %s: %v", containerName, err)
		return err
	}
	log.Printf("successfully removed container %s", containerName)
	return nil
}
