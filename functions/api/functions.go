package main

import (
	"log"
	"os/exec"
	"strings"
)

func PullImage(image string) error {
	cmd := exec.Command("nerdctl", "pull", image)
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func RunImage(image string, params ...string) (string, error) {
	param_str := strings.Join(params, " ")
	cmd := exec.Command("nerdctl", "run", "--rm", image, param_str)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(out), nil
}

func ImageStatus(image string) (string, error) {
	// run nerdctl inspect image and capture the output
	cmd := exec.Command("nerdctl", "inspect", image)
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return string(out), nil
}
