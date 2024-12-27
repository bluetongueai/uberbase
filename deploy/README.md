# Deploy CLI Tool

A command-line tool for deploying containerized applications using Podman with zero-downtime blue-green deployments.

## Features

- Zero-downtime deployments using blue-green strategy
- Multi-host deployment coordination
- Automatic health checking and rollback
- Volume and network management
- Registry authentication
- Backup management
- SSH-based remote deployment

## Installation

Download the latest release from the releases page or build from source:

```bash
go install github.com/bluetongueai/uberbase/deploy@latest
```

## Quick Start

1. Create a docker-compose.yml file for your services
2. Run the deploy command:

```bash
deploy -i ~/.ssh/id_rsa prod.example.com
```

## Usage

```
deploy [options] <ssh-host>

Arguments:
  <ssh-host>             SSH host to deploy to

Optional:
  -h, --help             Show this help message
  -f <file>              Path to docker-compose.yml (default: docker-compose.yml)
  --ssh-user <user>      SSH user (default: root)
  --ssh-port <port>      SSH port (default: 22)
  -i <keyfile>          SSH private key file
  --ssh-key-env <name>   Environment variable containing SSH key (default: SSH_PRIVATE_KEY)
  --registry <url>       Registry URL (default: docker.io)
  --registry-user <user> Registry username
  --registry-pass <pass> Registry password
```

## Examples

### Basic Deployment

Deploy using an SSH key file:
```bash
deploy prod.example.com --ssh-user deploy -i ~/.ssh/prod_key
```

Deploy using an SSH key from environment:
```bash
SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" deploy prod.example.com --ssh-user deploy
```

Deploy with a custom compose file:
```bash
deploy prod.example.com -f docker-compose.prod.yml
```

### Multi-Host Deployment

Deploy to multiple hosts:
```bash
deploy --hosts "host1.example.com,host2.example.com" --ssh-user deploy -i ~/.ssh/prod_key
```

### Private Registry

Deploy using a private registry:
```bash
deploy prod.example.com \
  --registry registry.example.com \
  --registry-user myuser \
  --registry-pass mypass
```

## Docker Compose Configuration

### Basic Service
```yaml
services:
  web:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - web_data:/var/www
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost/"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Backup Configuration
Enable automated backups by adding labels:

```yaml
services:
  database:
    image: postgres:13
    volumes:
      - db_data:/var/lib/postgresql/data
    labels:
      bluetongue.backup.enabled: "true"
      bluetongue.backup.schedule: "0 2 * * *"  # Daily at 2am
      bluetongue.backup.retention: "7d"        # Keep for 7 days
      bluetongue.backup.type: "volume"
```

## Health Checks

Health checks are crucial for zero-downtime deployments. Configure them in your docker-compose.yml:

```yaml
services:
  app:
    image: myapp:latest
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## Best Practices

1. Always configure health checks for your services
2. Use named volumes for persistent data
3. Enable backups for critical services
4. Use specific version tags for images
5. Test deployments in a staging environment first
6. Keep SSH keys secure and use non-root deploy users
7. Monitor deployment logs for issues

## Troubleshooting

### Common Issues

1. SSH Connection Failed
```bash
# Check SSH key permissions
chmod 600 ~/.ssh/id_rsa

# Test SSH connection
ssh -i ~/.ssh/id_rsa user@host
```

2. Health Check Failed
- Verify service is running correctly
- Check health check endpoint is accessible
- Review service logs for errors

3. Registry Authentication Failed
- Verify registry credentials
- Ensure registry URL is correct
- Check network connectivity to registry

### Rollback

To rollback to the previous version:

```bash
deploy --rollback prod.example.com -i ~/.ssh/prod_key
```

## Requirements

### Target Host
- Linux system with kernel 3.10+
- Minimum 2GB RAM
- Minimum 10GB available disk space
- Supported OS: Ubuntu, Debian, CentOS, RHEL, or Fedora

### Local System
- Go 1.16 or later (for building from source)
- SSH client
- Access to container registry

## Support

For issues and feature requests, please open an issue on the GitHub repository.
