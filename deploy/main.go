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
	composePath  string
	sshUser      string
	sshPort      int
	sshKeyFile   string
	sshKeyEnv    string
	registryURL  string
	regUser      string
	regPass      string
	hostsFlag    string
	rollbackFlag bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "deploy [flags] [hosts...]",
	Short: "A tool to deploy services from docker-compose.yml",
	Long: `A tool to deploy services from docker-compose.yml.

Examples:
  # Using SSH key file
  deploy prod.example.com --ssh-user deploy -i ~/.ssh/prod_key

  # Using SSH key from environment
  SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" deploy prod.example.com --ssh-user deploy

  # Using custom compose file
  deploy prod.example.com --ssh-user deploy -f docker-compose.prod.yml

  # Minimal usage
  deploy prod.example.com`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse hosts from flag or positional arguments
		var host string
		if hostsFlag != "" {
			host = hostsFlag
		} else if len(args) > 0 {
			host = args[0]
		} else {
			return fmt.Errorf("no hosts specified")
		}

		// Get SSH key from environment if specified and no key file provided
		logging.Logger.Debug("Determining SSH key")
		sshKeySource := bt_ssh.File
		if sshKeyFile == "" {
			logging.Logger.Debug("Using SSH key from environment")
			sshKeySource = bt_ssh.Environment
		}
		var sshKeyData string
		if sshKeyFile == "" {
			logging.Logger.Debug("No SSH key file provided, using SSH data from environment")
			sshKeyData = os.Getenv("SSH_PRIVATE_KEY")
		}
		if sshKeyData == "" && sshKeyFile == "" {
			return fmt.Errorf("either SSH key file (-i) or SSH key environment variable (SSH_PRIVATE_KEY) must be provided")
		}
		logging.Logger.Debug("Loading SSH key")
		sshKey := bt_ssh.SSHKey{
			Source: sshKeySource,
		}
		_, err := sshKey.Load()
		if err != nil {
			return fmt.Errorf("failed to load SSH key: %w", err)
		}

		// if not given a docker-compose.yml file, try to find one in the working directory
		if composePath == "" {
			logging.Logger.Debugf("No docker-compose.yml file provided, searching for one in the working directory")
			paths, err := filepath.Glob("docker-compose.yml")
			if err != nil {
				return fmt.Errorf("failed to find docker-compose.yml: %w", err)
			}
			if len(paths) == 0 {
				return fmt.Errorf("no docker-compose.yml file found in working directory")
			}
			composePath = paths[0]
		} else {
			logging.Logger.Debugf("Using docker-compose.yml file: %s", composePath)
		}

		localWorkDir := filepath.Dir(composePath)
		logging.Logger.Infof("Using local work directory: %s", localWorkDir)

		// load docker-compose.yml
		logging.Logger.Debugf("Loading docker-compose.yml")
		compose, err := containers.NewComposeProject(composePath, "uberbase-deploy")
		if err != nil {
			return fmt.Errorf("failed to load docker-compose.yml: %w", err)
		}

		logging.Logger.Debugf("Creating local executor")
		localExecutor := core.NewLocalExecutor()
		localExecutor.Verify()

		logging.Logger.Debugf("Creating remote executor")
		remoteExecutor, err := core.NewRemoteExecutor(bt_ssh.SSHConfig{
			Host: host,
			User: sshUser,
			Port: sshPort,
			Key:  sshKey,
		})
		if err != nil {
			return fmt.Errorf("failed to create remote executor: %w", err)
		}

		logging.Logger.Debugf("Creating deployer")
		deployer, err := deploy.NewDeployer(localExecutor, remoteExecutor, compose, localWorkDir, "~/uberbase-deploy")
		if err != nil {
			return fmt.Errorf("failed to create deployer: %w", err)
		}

		logging.Logger.Infof("Beginning deployment to %s", host)
		deployer.DeployProject()

		return nil
	},
}

func init() {
	// Define flags
	rootCmd.PersistentFlags().StringVarP(&composePath, "file", "f", "docker-compose.yml", "Path to docker-compose.yml")
	rootCmd.PersistentFlags().StringVar(&sshUser, "ssh-user", "root", "SSH user")
	rootCmd.PersistentFlags().IntVar(&sshPort, "ssh-port", 22, "SSH port")
	rootCmd.PersistentFlags().StringVarP(&sshKeyFile, "identity-file", "i", "", "SSH private key file")
	rootCmd.PersistentFlags().StringVar(&sshKeyEnv, "ssh-key-env", "SSH_PRIVATE_KEY", "Environment variable containing SSH key")
	rootCmd.PersistentFlags().StringVar(&registryURL, "registry", "", "Registry URL")
	rootCmd.PersistentFlags().StringVar(&regUser, "registry-user", "", "Registry username")
	rootCmd.PersistentFlags().StringVar(&regPass, "registry-pass", "", "Registry password")
	rootCmd.PersistentFlags().StringVar(&hostsFlag, "hosts", "", "Comma-separated list of hosts to deploy to")
	rootCmd.PersistentFlags().BoolVar(&rollbackFlag, "rollback", false, "Rollback deployment")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
