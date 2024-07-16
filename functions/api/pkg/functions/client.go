package functions

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/goombaio/namegenerator"
)

var seed = time.Now().UTC().UnixNano()
var nameGenerator = namegenerator.NewNameGenerator(seed)
var nameCounts = make(map[string]int)
var c client

type client struct {
	dockerPath        string
}

func newClient() (client, error) {
	client := client{}

	if c != client {
		return c, nil
	}

	log.Println("detecting container runtime")

	// check for docker
	path, err := exec.LookPath("docker")
	if err != nil {
		return client, fmt.Errorf("docker not found")
	}
	client.dockerPath = path

	log.Printf("using paths: docker=%s", client.dockerPath)

	c = client

	return client, nil
}

func (c client) command(bin string, args ...string) (string, string, error) {
	ctx := context.Background()
	var cmd *exec.Cmd
	log.Printf("running command %s %v", bin, args)
	cmd = exec.CommandContext(ctx, bin, args...)
	stdoutBuffer := &bytes.Buffer{}
	stderrBuffer := &bytes.Buffer{}
	cmd.Stdout = stdoutBuffer
	cmd.Stderr = stderrBuffer
	err := cmd.Run()
	if err != nil {
		return stdoutBuffer.String(), stderrBuffer.String(), err
	}
	return stdoutBuffer.String(), stderrBuffer.String(), nil
}

func (c client) docker(args ...string) (string, string, error) {
	return c.command(c.dockerPath, args...)
}

func (c client) dockerCompose(args ...string) (string, string, error) {
	cmdArgs := append([]string{"compose"}, args...)
	return c.command(c.dockerPath, cmdArgs...)
}

func (c client) Pull(imageName string, force bool) error {
	if force || !c.imageExists(imageName) {
		log.Printf("fetching docker image %s", imageName)
		_, _, err := c.docker("pull", imageName)
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
	_, _, err := c.docker("inspect", imageName)
	if err != nil {
		return false
	}
	return true
}

func (c client) Build(imageName, dockerfile string, context string) error {
	log.Printf("building docker image %s", imageName)
	_, _, err := c.docker("build", "-t", imageName, "-f", dockerfile, context)
	if err != nil {
		log.Printf("failed to build image %s: %v", imageName, err)
		return err
	}
	log.Printf("successfully built image %s", imageName)
	return nil
}

func (c client) Run(imageName string, params ...string) (string, string, error) {
	paramString := string.Join(" ", params)
	imageParams := append([]string{"run", "--rm", imageName}, paramString)
	stdout, stderr, err := c.docker(imageParams...)
	if err != nil {
		log.Printf("failed to run image %s: %v", imageName, err)
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}
