package main

import (
	"testing"
)

const testImage = "bluetongueai/functions-hello-world"

func removeImage() {
	// remove the test image using nerdctl
	_, err := Nerdctl("rmi", testImage)
	if err != nil {
		println("failed to remove image", err)
	}
}

func TestPullImage(t *testing.T) {
	removeImage()
	println("pulling image")
	err := PullImage("bluetongueai/functions-hello-world")
	if err != nil {
		t.Errorf("PullImage failed: %v", err)
	}
}

func TestRunImage(t *testing.T) {
	err := PullImage("bluetongueai/functions-hello-world")
	if err != nil {
		t.Errorf("PullImage failed: %v", err)
	}
	output, err := RunImage("my-image", "--param1", "value1", "--param2", "value2")
	if err != nil {
		t.Errorf("RunImage failed: %v", err)
	}
	if output != "Hello, world!" {
		t.Errorf("RunImage failed: expected output 'Hello, world!', got '%s'", output)
	}
}
