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
	limaPath    string
	limactlPath string
	nerdctlPath string
}

func newClient() (client, error) {
	client := client{}

	if c != client {
		return c, nil
	}

	log.Println("detecting container runtime")

	// check for lima
	path, err := exec.LookPath("lima")
	if err != nil {
		return client, fmt.Errorf("lima not found")
	}
	client.limaPath = path

	// check for limactl
	path, err = exec.LookPath("limactl")
	if err != nil {
		return client, fmt.Errorf("limactl not found")
	}
	client.limactlPath = path

	// check for nerdctl.lima
	path, err = exec.LookPath("nerdctl.lima")
	if err != nil {
		return client, fmt.Errorf("nerdctl.lima not found")
	}
	client.nerdctlPath = path

	log.Printf("using paths: lima=%s limactl=%s nerdctl=%s", client.limaPath, client.limactlPath, client.nerdctlPath)

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

func (c client) lima(args ...string) (string, string, error) {
	return c.command(c.limaPath, args...)
}

func (c client) limactl(args ...string) (string, string, error) {
	return c.command(c.limactlPath, args...)
}

func (c client) nerdctl(args ...string) (string, string, error) {
	return c.command(c.nerdctlPath, args...)
}

func (c client) Pull(imageName string, force bool) error {
	if force || !c.imageExists(imageName) {
		log.Printf("fetching containerd image %s", imageName)
		_, _, err := c.nerdctl("pull", imageName)
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
	_, _, err := c.nerdctl("inspect", imageName)
	if err != nil {
		return false
	}
	return true
}

func (c client) Build(imageName, dockerfile string, context string) error {
	log.Printf("building containerd image %s", imageName)
	_, _, err := c.nerdctl("build", "-t", imageName, "-f", dockerfile, context)
	if err != nil {
		log.Printf("failed to build image %s: %v", imageName, err)
		return err
	}
	log.Printf("successfully built image %s", imageName)
	return nil
}

func (c client) Run(imageName string, params ...string) (string, string, error) {
	imageParams := append([]string{"run", imageName}, params...)
	stdout, stderr, err := c.nerdctl(imageParams...)
	if err != nil {
		log.Printf("failed to run image %s: %v", imageName, err)
		return "", "", err
	}
	return stdout, stderr, nil
}
