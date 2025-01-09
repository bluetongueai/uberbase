package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bluetongueai/uberbase/deploy/pkg/containers"
	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/deploy"
	"github.com/bluetongueai/uberbase/deploy/pkg/logging"
	bt_ssh "github.com/bluetongueai/uberbase/deploy/pkg/ssh"
	"github.com/spf13/cobra"
)

var (
	// Command line flags
	composePath string
	sshUser     string
	sshPort     int
	sshKeyFile  string
	sshKeyEnv   string
	registryURL string
	regUser     string
	regPass     string
	debug       bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deploy [flags] host",
	Short: "A tool to deploy services from docker-compose.yml",
	Long: `A tool to deploy services from docker-compose.yml.

Examples:
  # Using SSH key file
  deploy prod.example.com --ssh-user deploy -i ~/.ssh/prod_key

  # Using SSH key from environment
  SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" deploy prod.example.com --ssh-user deploy

  # Minimal usage
  deploy prod.example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if debug {
			logging.SetDebugLevel()
		}

		// Parse hosts from flag or positional arguments
		var host string
		if len(args) > 0 {
			host = args[0]
		} else {
			return fmt.Errorf("no hosts specified")
		}

		// Get SSH key configuration
		sshKeySource := bt_ssh.File
		var sshKeyData string
		if sshKeyFile == "" {
			sshKeySource = bt_ssh.Environment
			sshKeyData = os.Getenv("SSH_PRIVATE_KEY")
			logging.Logger.Debug("Using SSH key from environment")
		} else {
			logging.Logger.Debug("Using SSH key from filepath", sshKeyFile)
		}

		if sshKeyData == "" && sshKeyFile == "" {
			return fmt.Errorf("either SSH key file (-i) or SSH key environment variable (SSH_PRIVATE_KEY) must be provided")
		}

		sshKey := bt_ssh.NewSSHKey(sshKeySource, sshKeyEnv, sshKeyFile)
		if _, err := sshKey.Load(); err != nil {
			return fmt.Errorf("failed to load SSH key: %w", err)
		}
		logging.Logger.Debug("SSH key loaded successfully", "source", sshKeySource)

		// Locate docker-compose file
		if composePath == "" {
			paths, err := filepath.Glob("docker-compose.yml")
			if err != nil {
				return fmt.Errorf("failed to find docker-compose.yml: %w", err)
			}
			if len(paths) == 0 {
				return fmt.Errorf("no docker-compose.yml file found in working directory")
			}
			composePath = paths[0]
			logging.Logger.Debug("Found docker-compose.yml in current directory", "path", composePath)
		}
		localWorkDir := filepath.Dir(composePath)

		logging.LogKeyValues("Initializing deployment", map[string]string{
			"host":           host,
			"local workdir":  "\033[33m" + localWorkDir + "\033[0m",
			"remote workdir": "\033[34m" + "~/uberbase-deploy" + "\033[0m",
		})

		// Load docker-compose configuration
		compose, err := containers.NewComposeProject(composePath, "uberbase-deploy")
		if err != nil {
			return fmt.Errorf("failed to load docker-compose.yml: %w", err)
		}
		logging.Logger.Debug("Loaded compose configuration ",
			"services", compose.Project.Services,
			"project", compose.Project.Name)

		// Initialize executors
		localExecutor := core.NewLocalExecutor()
		remoteExecutor, err := core.NewRemoteExecutor(bt_ssh.SSHConfig{
			Host: host,
			User: sshUser,
			Port: sshPort,
			Key:  *sshKey,
		})
		if err != nil {
			return fmt.Errorf("failed to create remote executor: %w", err)
		}

		// Create and run deployer
		deployer, err := deploy.NewDeployer(localExecutor, remoteExecutor, compose, localWorkDir, "~/uberbase-deploy")
		if err != nil {
			return fmt.Errorf("failed to create deployer: %w", err)
		}

		logging.Logger.Info("Starting deployment to", "host", host)
		if err := deployer.DeployProject(); err != nil {
			logging.Logger.Error("Deployment failed", "error", err)
			return err
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	// Initialize logging
	logging.InitLogging()

	// Define flags
	rootCmd.PersistentFlags().StringVarP(&composePath, "file", "f", "", "Path to docker-compose.yml (default: ./docker-compose.yml)")
	rootCmd.PersistentFlags().StringVar(&sshUser, "ssh-user", "root", "SSH user")
	rootCmd.PersistentFlags().IntVar(&sshPort, "ssh-port", 22, "SSH port")
	rootCmd.PersistentFlags().StringVarP(&sshKeyFile, "identity-file", "i", "", "SSH private key file")
	rootCmd.PersistentFlags().StringVar(&sshKeyEnv, "ssh-key-env", "SSH_PRIVATE_KEY", "Environment variable containing SSH key")
	rootCmd.PersistentFlags().StringVar(&registryURL, "registry", "", "Registry URL")
	rootCmd.PersistentFlags().StringVar(&regUser, "registry-user", "", "Registry username")
	rootCmd.PersistentFlags().StringVar(&regPass, "registry-pass", "", "Registry password")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
