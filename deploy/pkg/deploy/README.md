err = coordinator.DeployCompose(config, "v1.0")

### Backup Configuration

// Configure backup in compose file
services:
  myapp:
    image: nginx:latest
    labels:
      bluetongue.backup.enabled: "true"
      bluetongue.backup.schedule: "0 2 * * *"  # Daily at 2am
      bluetongue.backup.retention: "7d"         # Keep for 7 days
# Deploy Package

The deploy package provides a comprehensive deployment system for containerized applications using Podman. It implements blue-green deployments, state management, backup handling, and health monitoring.
      bluetongue.backup.type: "volume"

### State Management

// Get deployment state
state, err := deployer.StateManager().Load()

// Check service version
serviceState := state.Services["myapp"]
currentVersion := serviceState.BlueVersion

### Health Monitoring

// Configure health checks
service.Container.Healthcheck = &podman.Healthcheck{
    Test: []string{"CMD", "curl", "-f", "http://localhost/health"},

## Core Components

### 1. Deployer
The main orchestrator that coordinates deployments:
- Manages container lifecycle
- Handles blue-green deployments
- Coordinates backups
- Monitors service health
- Routes traffic between versions

### 2. Service Manager
Handles service operations:
- Creates and removes containers
- Manages volumes and networks
- Handles image pulling
- Migrates data between versions
    Interval: "30s",
    Timeout: "10s",
    Retries: 3,
}

## Key Features

1. Blue-Green Deployments
- Zero-downtime deployments
- Automatic health checking
- Gradual traffic shifting
- Automatic rollback on failure

2. State Management
- Persistent deployment state
- Distributed locking
- Transaction logging
- Rollback information

3. Backup System
- Scheduled backups
- Configurable retention

### 3. State Manager
Maintains deployment state:
- Persists deployment configurations
- Manages deployment locks
- Tracks service versions
- Records transaction history

### 4. Traffic Manager
Controls request routing:
- Updates load balancer weights
- Manages gradual traffic shifting
- Handles service discovery labels

### 5. Backup Manager
Handles service backups:
- Schedules regular backups
- Manages backup retention
- Space validation
- Pre-deployment backups

4. Health Monitoring
- Container health checks
- Custom health check commands
- Configurable retry policies
- Automatic failure handling

5. Multi-Host Deployment
- Coordinated deployments
- Distributed locking
- State synchronization
- Parallel execution

## Best Practices

1. Always configure health checks for your services
2. Use persistent volumes for stateful services
- Validates backup space
- Handles backup restoration

## Usage Examples

### Basic Service Deployment

// Create a new deployer
ssh := core.NewSSHConnection("host", "user", "key")
deployer := NewDeployer(ssh, "/var/lib/deploy", podman.RegistryConfig{})

// Define a service
service := &Service{
3. Enable backups for critical data
4. Set appropriate resource limits
5. Use meaningful version tags
6. Monitor deployment metrics
7. Configure appropriate backup retention periods
8. Test rollback procedures

## Error Handling

The package provides comprehensive error handling:
- Deployment failures trigger automatic rollbacks
- Failed health checks prevent traffic shifting
- Backup failures are logged but non-blocking
- Lock timeouts prevent stuck deployments
    Name: "myapp",
    Image: podman.ParseImageRef("nginx:latest"),
    Container: &podman.Container{
        Ports: []string{"80:80"},
        Healthcheck: &podman.Healthcheck{
            Test: []string{"CMD", "curl", "-f", "http://localhost"},
        },
    },
- Transaction logs track failure points

## Metrics and Monitoring

The package collects various metrics:
- Deployment duration
- Health check results
- Backup sizes and durations
- Traffic distribution
- Resource usage

These metrics can be accessed through the DeploymentMetrics interface.
