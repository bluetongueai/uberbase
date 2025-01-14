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
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	shuttingDown := false
	go func() {
		<-quit
		if shuttingDown {
			return
		}
		shuttingDown = true
		log.Println("Shutting down Uberbase...")
		stopCmd := exec.Command("./bin/stop")
		stopCmd.Stdout = os.Stdout
		stopCmd.Stderr = os.Stderr
		if err := stopCmd.Run(); err != nil {
			log.Printf("Error shutting down Uberbase: %v", err)
		}
		os.Exit(0)
	}()
}

func main() {
	setupSignalHandling()
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
