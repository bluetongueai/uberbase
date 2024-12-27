package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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
	// Set custom usage before defining flags
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n", usage)
	}

	var (
		sshUser     = flag.String("ssh-user", "root", "SSH user")
		sshPort     = flag.Int("ssh-port", 22, "SSH port")
		sshKeyFile  = flag.String("i", "", "SSH private key file")
		sshKeyEnv   = flag.String("ssh-key-env", "SSH_PRIVATE_KEY", "Environment variable containing SSH key")
		registryURL = flag.String("registry", "docker.io", "Registry URL")
		regUser     = flag.String("registry-user", "", "Registry username")
		regPass     = flag.String("registry-pass", "", "Registry password")
		composePath = flag.String("f", "docker-compose.yml", "Path to docker-compose.yml")
	)

	// Find the ssh-host argument and remove it from os.Args
	var sshHost string
	var newArgs []string
	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "-") {
			newArgs = append(newArgs, arg)
			// If this flag expects a value, add the next argument too
			if arg == "-i" || arg == "-f" || strings.HasPrefix(arg, "--") {
				if i+1 < len(os.Args) {
					i++
					newArgs = append(newArgs, os.Args[i])
				}
			}
		} else if sshHost == "" {
			sshHost = arg
		}
	}

	// Reconstruct os.Args with the program name and processed args
	os.Args = append([]string{os.Args[0]}, newArgs...)

	// Parse flags
	flag.Parse()

	// Validate we got the ssh-host
	if sshHost == "" {
		core.Logger.Fatal("Exactly one argument (ssh-host) is required")
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

	// Create deployer with all configuration
	ssh, err := core.NewSession(core.SSHConfig{
		Host: sshHost,
		User: *sshUser,
		Port: *sshPort,
	})
	if err != nil {
		log.Fatalf("Failed to create SSH session: %v", err)
	}
	defer ssh.Close()

	registryConfig := podman.RegistryConfig{
		Host:     *registryURL,
		Username: *regUser,
		Password: *regPass,
	}

	// Generate version string based on current time
	version := time.Now().UTC().Format("20060102-150405") // Will output like "20240319-153022"

	deployer := deploy.NewDeployer(ssh, *composePath, registryConfig)

	if err := deployer.DeployCompose(composeConfig, version); err != nil {
		log.Fatalf("Deployment failed: %v", err)
	}
}
