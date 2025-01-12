package main

import (
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func getContainerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "container [container-args]",
		Short: "Manages containers, images, and volumes",
		Long:  `Manage containers, images, and volumes.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}

			// Prepare the podman command
			podmanCmd := exec.Command("podman", args...)
			podmanCmd.Stdout = os.Stdout
			podmanCmd.Stderr = os.Stderr
			podmanCmd.Stdin = os.Stdin

			// Run the podman command
			if err := podmanCmd.Run(); err != nil {
				return err
			}

			return nil
		},
	}
}
