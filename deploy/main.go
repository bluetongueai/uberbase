package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
	"github.com/bluetongueai/uberbase/deploy/pkg/deploy"
	"github.com/bluetongueai/uberbase/deploy/pkg/podman"
)

const usage = `Usage: deploy [options] <ssh-host>

A tool to deploy services from docker-compose.yml.

Arguments:
  <ssh-host>             SSH host to deploy to

Optional:
  -h, --help             Show this help message
  -f <file>              Path to docker-compose.yml (default: docker-compose.yml)
  --ssh-user <user>      SSH user (default: root)
  --ssh-port <port>      SSH port (default: 22)
  -i <keyfile>           SSH private key file
  --ssh-key-env <name>   Environment variable containing SSH key (default: SSH_PRIVATE_KEY)
  --registry <url>       Registry URL (default: docker.io)
  --registry-user <user> Registry username
  --registry-pass <pass> Registry password

Examples:
  # Using SSH key file
  deploy prod.example.com --ssh-user deploy -i ~/.ssh/prod_key

  # Using SSH key from environment
  SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" deploy prod.example.com --ssh-user deploy

  # Using custom compose file
  deploy prod.example.com --ssh-user deploy -f docker-compose.prod.yml

  # Minimal usage
  deploy prod.example.com
`

func main() {
	var (
		composePath = flag.String("f", "docker-compose.yml", "Path to docker-compose.yml")
		sshUser     = flag.String("ssh-user", "root", "SSH user")
		sshPort     = flag.Int("ssh-port", 22, "SSH port")
		sshKeyFile  = flag.String("i", "", "SSH private key file")
		sshKeyEnv   = flag.String("ssh-key-env", "SSH_PRIVATE_KEY", "Environment variable containing SSH key")
		registryURL = flag.String("registry", "", "Registry URL")
		regUser     = flag.String("reg-user", "", "Registry username")
		regPass     = flag.String("reg-pass", "", "Registry password")
		hostsFlag   = flag.String("hosts", "", "Comma-separated list of hosts to deploy to")
		rollback    = flag.Bool("rollback", false, "Rollback deployment")
	)

	flag.Usage = func() {
		fmt.Print(usage)
	}
	flag.Parse()

	// Parse hosts from flag or positional argument
	var hosts []string
	if *hostsFlag != "" {
		hosts = strings.Split(*hostsFlag, ",")
	} else if flag.NArg() > 0 {
		// Backward compatibility: single host as positional argument
		hosts = []string{flag.Arg(0)}
	} else {
		log.Fatal("No hosts specified")
	}

	// Read docker-compose.yml
	composeConfig, err := deploy.ParseComposeFile(*composePath)
	if err != nil {
		log.Fatalf("Failed to read %s: %v", *composePath, err)
	}

	// Get SSH key from environment if specified and no key file provided
	var sshKeyData string
	if *sshKeyFile == "" {
		sshKeyData = os.Getenv(*sshKeyEnv)
		if sshKeyData == "" {
			sshKeyData = os.Getenv("SSH_PRIVATE_KEY")
		}
		if sshKeyData == "" && *sshKeyFile == "" {
			core.Logger.Fatal("Either SSH key file (-i) or SSH key environment variable (SSH_PRIVATE_KEY) must be provided")
		}
	}

	registryConfig := podman.RegistryConfig{
		Host:     *registryURL,
		Username: *regUser,
		Password: *regPass,
	}

	// Generate version string based on current time
	version := time.Now().UTC().Format("20060102-150405") // Will output like "20240319-153022"

	// Add after composePath declaration
	workDir := filepath.Dir(*composePath)

	// Create deployer for each host
	deployers := make([]*deploy.Deployer, len(hosts))
	for i, host := range hosts {
		ssh, err := core.NewSession(core.SSHConfig{
			Host: host,
			User: *sshUser,
			Port: *sshPort,
		})
		if err != nil {
			log.Fatalf("Failed to create SSH session for %s: %v", host, err)
		}
		defer ssh.Close()

		deployers[i] = deploy.NewDeployer(ssh, workDir, registryConfig)
	}

	// Create coordinator with placement
	coordinator := deploy.NewDeploymentCoordinator(deployers)

	if err := coordinator.DeployCompose(composeConfig, version); err != nil {
		log.Printf("Deployment failed: %v", err)
		log.Println("Attempting rollback...")

		if *rollback {
			if err := coordinator.Rollback(composeConfig); err != nil {
				log.Fatalf("Failed to rollback: %v", err)
			}
			return
		}
		os.Exit(1)
	}
}
