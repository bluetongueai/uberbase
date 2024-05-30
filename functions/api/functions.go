package main

import (
	"log"
	"os/exec"
	"strings"
)

func Nerdctl(params ...string) (string, error) {
	// if the platform is linux, use nerdctl
	// if the platform is macos, use lima nerdctl
	command := "nerdctl"
	_, err := exec.LookPath("lima")
	if err == nil {
		command = "lima"
	}
	param_str := strings.Join(params, " ")
	if command == "lima" {
		param_str = strings.Join(append([]string{"nerdctl"}, params...), " ")
	}
	param_arr := strings.Split(param_str, " ")
	println("running", command, param_str)
	cmd := exec.Command(command, param_arr...)
	out, err := cmd.Output()
	if err != nil {
		println("output:", string(out))
		println("got err", err.Error())
		log.Fatal(err)
		return "", err
	}
	return string(out), nil
}

func PullImage(image string) error {
	_, err := Nerdctl("pull", image)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func RunImage(image string, params ...string) (string, error) {
	param_str := strings.Join(params, " ")
	out, err := Nerdctl("run", "--rm", image, param_str)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return out, nil
}

func ImageStatus(image string) (string, error) {
	// run nerdctl inspect image and capture the output
	out, err := Nerdctl("inspect", image)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return out, nil
}
