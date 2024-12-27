package core

import (
	"fmt"
	"os/exec"
	"strings"
)

func GetCurrentRepoURL() (string, error) {
	output, err := runLocalCommand("git config --get remote.origin.url")
	if err != nil {
		return "", fmt.Errorf("failed to get git remote: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func runLocalCommand(command string) (string, error) {
	cmd := exec.Command("sh", "-c", command)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
