package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"
)

func getStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start the Uberbase platform",
		Long:  `Configures and starts the Uberbase platform.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptPath := filepath.Join(".", "bin", "start")
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				return fmt.Errorf("start bin not found at %s", scriptPath)
			}

			startCmd := exec.Command(scriptPath)
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr
			startCmd.Stdin = os.Stdin

			// Set process group to ensure signals are propagated
			startCmd.SysProcAttr = &syscall.SysProcAttr{
				Setpgid: false, // Allow signal propagation from parent
			}

			return startCmd.Run()
		},
	}
}

func getStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stops the Uberbase platform",
		Long:  `Stops the Uberbase platform.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			scriptPath := filepath.Join(".", "bin", "stop")
			if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
				return fmt.Errorf("stop bin not found at %s", scriptPath)
			}

			stopCmd := exec.Command(scriptPath)
			stopCmd.Stdout = os.Stdout
			stopCmd.Stderr = os.Stderr
			stopCmd.Stdin = os.Stdin

			return stopCmd.Run()
		},
	}
}
