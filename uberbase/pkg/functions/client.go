package functions

import (
	"bytes"
	"fmt"
	"os/exec"
	"time"

	"github.com/bluetongueai/uberbase/uberbase/pkg/logging"
	"github.com/goombaio/namegenerator"
)

var seed = time.Now().UTC().UnixNano()
var nameGenerator = namegenerator.NewNameGenerator(seed)
var nameCounts = make(map[string]int)
var c client

type client struct {
	binPath string
}

func newClient() (client, error) {
	client := client{}

	if c != client {
		return c, nil
	}

	logging.Logger.Info("Detecting container runtime")

	// check for container manager runtime
	path, err := exec.LookPath("podman")
	if err != nil {
		return client, fmt.Errorf("podman not found")
	}
	client.binPath = path

	logging.Logger.Infof("Using paths: container runtime=%s", client.binPath)

	c = client

	return client, nil
}

func (c client) command(bin string, args ...string) (string, string, error) {
	var cmd *exec.Cmd
	logging.Logger.Infof("Running command %s %v", bin, args)
	cmd = exec.Command(bin, args...)
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

func (c client) container(args ...string) (string, string, error) {
	return c.command(c.binPath, args...)
}

func (c client) containerCompose(args ...string) (string, string, error) {
	// cmdArgs := append([]string{"compose"}, args...)
	composePath := c.binPath + "-compose"
	return c.command(composePath, args...)
}

func (c client) Pull(imageName string, force bool) (string, string, error) {
	if force || !c.imageExists(imageName) {
		logging.Logger.Infof("Fetching container image %s", imageName)
		stdout, stderr, err := c.container("pull", imageName)
		if err != nil {
			logging.Logger.Errorf("Failed to pull image %s: %v", imageName, err)
			return stdout, stderr, err
		}
		logging.Logger.Infof("Successfully pulled image %s", imageName)
		return stdout, stderr, nil
	}
	return "image already exists, force not specified", "", nil
}

func (c client) imageExists(imageName string) bool {
	_, _, err := c.container("inspect", imageName)
	if err != nil {
		return false
	}
	return true
}

func (c client) Build(imageName, containerfile string, context string) error {
	logging.Logger.Infof("Building container image %s", imageName)
	_, _, err := c.container("build", "-t", imageName, "-f", containerfile, context)
	if err != nil {
		logging.Logger.Errorf("Failed to build image %s: %v", imageName, err)
		return err
	}
	logging.Logger.Infof("Successfully built image %s", imageName)
	return nil
}

func (c client) Run(imageName string, detatch bool, env map[string]string, params ...string) (string, string, error) {
	// imageParams := append([]string{"run", "--rm", "-i", imageName}, params...)
	imageParams := []string{"run", "--env-file", ".env", "--rm", "-i", "--net", "uberbase_net"}
	if detatch {
		imageParams = append(imageParams, "-d")
	}
	for k, v := range env {
		imageParams = append(imageParams, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	imageParams = append(append(imageParams, imageName), params...)
	stdout, stderr, err := c.container(imageParams...)
	if err != nil {
		logging.Logger.Errorf("Failed to run image %s: %v", imageName, err)
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}

func (c client) Stop(containerId string) (string, string, error) {
	// imageParams := append([]string{"run", "--rm", "-i", imageName}, params...)
	stdout, stderr, err := c.container("stop", containerId)
	if err != nil {
		logging.Logger.Errorf("Failed to stop container %s: %v", containerId, err)
		return stdout, stderr, err
	}
	return stdout, stderr, nil
}
