package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "uberbase",
	Short: "Uberbase CLI tool",
	Long:  `A command line tool for managing Uberbase deployments and services.`,
}

func init() {
	// Add all subcommands
	rootCmd.AddCommand(getServeCmd())
	rootCmd.AddCommand(getDeployCmd())
	rootCmd.AddCommand(getContainerCmd())
	rootCmd.AddCommand(getStartCmd())
	rootCmd.AddCommand(getStopCmd())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
