# Podman Package

The podman package provides a Go interface for managing containers, images, volumes, networks, and pods using Podman through SSH connections. It offers a high-level abstraction over Podman commands with robust error handling and logging.

## Core Components

### Container Management
The package provides container lifecycle management through the `ContainerManager`:

// Create a container manager
ssh := core.NewSession(sshConfig)
manager := podman.NewContainerManager(ssh)

// Create and run a container
container := podman.Container{
    Name:  "web",
    Image: "nginx:latest",
    Ports: []string{"80:80"},
    Volumes: []string{"nginx_data:/var/www"},
    Environment: map[string]string{
        "NGINX_HOST": "example.com",
    },
    Healthcheck: &podman.Healthcheck{
        Test:     []string{"CMD", "curl", "-f", "http://localhost/"},
        Interval: "30s",
        Timeout:  "10s",
        Retries:  3,
    },
}

err := manager.Create(container)

### Volume Management
The `VolumeManager` handles persistent storage:

// Create a volume manager
volManager := podman.NewVolumeManager(ssh)

// Create a volume
err := volManager.EnsureVolume("mydata")

// Handle bind mounts and named volumes
volumes := []string{
    "mydata:/app/data",
    "/host/path:/container/path:ro"
}
err = volManager.EnsureVolumes(volumes)

### Network Management
The `NetworkManager` handles container networking:

// Create a network manager
netManager := podman.NewNetworkManager(ssh)

// Create a network
err := netManager.EnsureNetwork("backend", true) // true for internal network

// Connect container to networks
err = netManager.ConnectContainer("myapp", []string{"frontend", "backend"})

### Image Management
The `RegistryClient` handles image operations:

// Create a registry client
registry := podman.New(ssh, podman.RegistryConfig{
    Host:     "registry.example.com",
    Username: "user",
    Password: "pass",
})

// Pull an image
imageRef := podman.ImageRef{
    Name: "myapp",
    Tag:  "v1.0",
}
err := registry.PullImage(imageRef)

### Pod Management
The `PodManager` handles pod operations:

// Create a pod manager
podManager := podman.NewPodManager(ssh)

// Create a pod
pod := podman.Pod{
    Name:     "web-stack",
    Networks: []string{"frontend"},
    Ports:    []string{"80:80"},
}
err := podManager.Create(pod)

## Features

### Health Checking
- Built-in container health checking
- Configurable health check intervals and retries
- Dependency health monitoring

### Volume Management
- Named volume creation and management
- Bind mount support with options (ro, rw, etc.)
- SELinux context handling
- Volume backup and restore

### Network Management
- Network creation and cleanup
- Multiple network attachment
- Internal network support
- DNS configuration

### Container Features
- Environment variable management
- Port mapping
- Volume mounting
- Health checks
- Resource limits
- Network configuration
- Security options
- Dependency management

### Registry Operations
- Image pulling and pushing
- Tag management
- Registry authentication
- Image existence checking
- Image digest retrieval

## Best Practices

1. Error Handling
Always check returned errors:

err := manager.Create(container)
if err != nil {
    // Handle error appropriately
    log.Printf("Failed to create container: %v", err)
    return err
}

2. Resource Cleanup
Use defer for cleanup operations:

defer manager.Remove(container.Name)

3. Health Checks
Implement health checks for production containers:

container.Healthcheck = &podman.Healthcheck{
    Test:     []string{"CMD", "curl", "-f", "http://localhost/health"},
    Interval: "30s",
    Timeout:  "10s",
    Retries:  3,
}

4. Network Isolation
Use internal networks for backend services:

err := netManager.EnsureNetwork("backend", true)
err = netManager.ConnectContainer("db", []string{"backend"})

5. Volume Persistence
Use named volumes for persistent data:

volumes := []string{
    "db_data:/var/lib/postgresql/data",
    "redis_data:/data"
}
err = volManager.EnsureVolumes(volumes)

## Error Handling

The package provides detailed error messages and logging:

- Container creation failures include cleanup operations
- Network connection errors include retry logic
- Volume mounting errors include detailed context
- Registry operations handle authentication failures

## Dependencies

- Podman installed on the remote system
- SSH access to the remote system
- Appropriate permissions for Podman operations

## Thread Safety

The package is designed to be thread-safe for concurrent operations on different resources. However, operations on the same resource (e.g., the same container) should be synchronized by the caller.

## Logging

The package uses structured logging through the core.Logger interface:

core.Logger.Info("Creating container")
core.Logger.Error("Failed to create container", err)

## Examples

### Complete Container Deployment

manager := podman.NewContainerManager(ssh)
volManager := podman.NewVolumeManager(ssh)
netManager := podman.NewNetworkManager(ssh)

// Create network
err := netManager.EnsureNetwork("app-network", false)
if err != nil {
    return err
}

// Ensure volume
err = volManager.EnsureVolume("app-data")
if err != nil {
    return err
}

// Create container
container := podman.Container{
    Name:  "app",
    Image: "myapp:latest",
    Ports: []string{"8080:8080"},
    Volumes: []string{"app-data:/app/data"},
    Networks: []string{"app-network"},
    Environment: map[string]string{
        "DB_HOST": "db",
        "DB_PORT": "5432",
    },
    Healthcheck: &podman.Healthcheck{
        Test:     []string{"CMD", "curl", "-f", "http://localhost:8080/health"},
        Interval: "30s",
        Timeout:  "10s",
        Retries:  3,
    },
}

err = manager.Create(container)
if err != nil {
    return err
}

### Database with Persistent Storage

// Create volume
err := volManager.EnsureVolume("postgres-data")
if err != nil {
    return err
}

// Create database container
container := podman.Container{
    Name:  "db",
    Image: "postgres:13",
    Volumes: []string{"postgres-data:/var/lib/postgresql/data"},
    Environment: map[string]string{
        "POSTGRES_PASSWORD": "secret",
        "POSTGRES_DB": "myapp",
    },
    Healthcheck: &podman.Healthcheck{
        Test:     []string{"CMD-SHELL", "pg_isready -U postgres"},
        Interval: "10s",
        Timeout:  "5s",
        Retries:  5,
    },
}

err = manager.Create(container)
