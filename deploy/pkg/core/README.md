# Core Package

The core package provides fundamental functionality for remote system management and container orchestration. It includes SSH connectivity, Podman installation, logging, and Git operations.

## Components

### SSH Connection (rexec.go)

Manages SSH connections and remote command execution:

```go
// Create a new SSH session
config := SSHConfig{
    Host:    "example.com",
    User:    "username",
    Port:    22,
    KeyFile: "/path/to/key",
}

conn, err := NewSession(config)
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

// Execute remote commands
output, err := conn.Exec("ls -l")
```

### Podman Installation (podman.go)

Handles automated installation of Podman and Podman Compose on remote systems:

```go
installer := NewPodmanInstaller(sshConn)

// Install Podman and dependencies
err := installer.EnsureInstalled()
if err != nil {
    log.Fatal(err)
}
```

Key features:
- OS detection and package manager selection
- System requirements validation
- Automatic dependency installation
- Installation retry logic
- Cleanup on failure

### Git Operations (git.go)

Provides Git-related functionality:

```go
// Get repository URL
url, err := GetCurrentRepoURL()
if err != nil {
    log.Fatal(err)
}
```

### Logging (logging.go)

Centralized logging using logrus:

```go
// Initialize logging
InitLogging()

// Use logger
Logger.Info("Starting deployment")
Logger.Debug("Detailed information")
Logger.Error("Something went wrong")
```

### Mock SSH (mock_ssh.go)

Testing utilities for SSH operations:

```go
// Create mock SSH server
mock := NewMockSSH("localhost:2222")
defer mock.Close()

// Configure mock responses
mock.SetReturnString("expected output")
```

## System Requirements

For Podman installation, target systems must have:
- Linux kernel version 3.10+
- Minimum 2GB RAM
- Minimum 10GB available disk space
- Supported OS: Ubuntu, Debian, CentOS, RHEL, or Fedora

## Error Handling

The package implements comprehensive error handling for:
- SSH connection failures
- Command execution errors
- System requirement validation
- OS compatibility checks
- Installation failures

## Testing

Run the tests using:

```bash
go test ./...
```

The package includes extensive tests using mock SSH servers for reliable testing of remote operations. 
