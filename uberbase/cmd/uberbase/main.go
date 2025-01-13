package main

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

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

// set up signal handling
func setupSignalHandling() {
	log.Println("Setting up signal handling...")
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-quit
		log.Printf("Received signal: %v", sig)

		log.Println("Running cleanup...")
		stopCmd := exec.Command("./bin/stop")
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		if err := stopCmd.Run(); err != nil {
			log.Printf("Cleanup error: %v", err)
		}

		os.Exit(0)
	}()
}

func main() {
	log.Println("Starting Uberbase CLI...")
	setupSignalHandling()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
