package cli

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/spf13/cobra"
)

// Root command
var rootCmd = &cobra.Command{
	Use:   "uberbase",
	Short: "A full-featured platform-in-a-box",
	Long:  `Uberbase provides a complete development and deployment environment.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Add debug flag to root command
	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug output")

	// Add commands
	rootCmd.AddCommand(newStartCmd())
	rootCmd.AddCommand(newStopCmd())
	rootCmd.AddCommand(newDeployCmd())
	rootCmd.AddCommand(newComposeCmd())
	rootCmd.AddCommand(newPodmanCmd())
	rootCmd.AddCommand(newVaultCmd())
}

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start Uberbase services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCommand("./bin/start", args)
		},
	}
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop all services",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCommand("./bin/stop", args)
		},
	}
}

func newDeployCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "deploy",
		Short: "Deploy containers into an SSH-able host",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCommand("./bin/deploy", args)
		},
	}
}

func newComposeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "compose",
		Short: "Manage containers using Podman Compose",
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCommand("podman-compose", args)
		},
	}
}

// Handle all other podman commands
func newPodmanCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "podman",
		Short:              "Execute podman commands directly",
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return executeCommand("podman", args)
		},
	}
}

// Add this new function after the other command functions
func newVaultCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vault",
		Short: "Execute Vault commands inside the Vault container",
		Long: `Execute HashiCorp Vault commands inside the running Vault container.
Examples:
  uberbase vault status
  uberbase vault token create
  uberbase vault secrets list`,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Construct command to execute vault CLI inside the container
			vaultArgs := append([]string{"exec", "-it", "vault", "vault"}, args...)
			return executeCommand("podman", vaultArgs)
		},
	}
}

// executeCommand handles running external commands and properly forwarding I/O and exit codes
func executeCommand(binary string, args []string) error {
	cmd := exec.Command(binary, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return err
	}
	return nil
}
