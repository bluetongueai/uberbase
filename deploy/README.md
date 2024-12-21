# Deploy Package Documentation

## builder.go

The builder package provides functionality for building container images using Podman.

### Types

#### BuildOptions
Configures how container images are built. Key configuration groups include:

```go
type BuildOptions struct {
    // Core Build Configuration
    File         string            // Path to Dockerfile or URL
    ContextPath  string            // Root directory for the build context
    Tag          string            // Tag for the built image

    // Build Arguments and Environment
    BuildArgs     map[string]string // Build-time variables (--build-arg)
    BuildArgFile  string           // File containing build arguments
    Environment   map[string]string // Environment variables for build process
    Unsetenv     []string         // Environment variables to unset

    // Multi-stage Build Controls
    BuildContexts     map[string]string // Named build contexts
    FromStage        string            // Stage to build from in multi-stage builds
    Target           string            // Target stage to build in multi-stage builds
    SkipUnusedStages bool             // Skip building unused stages

    // Cache Configuration
    NoCache     bool                // Disable build cache
    CacheFrom   []string           // External cache sources
    CacheTo     []string           // External cache destinations
    CacheTTL    string            // Cache time-to-live

    // Image Metadata
    Format         string          // Output format (docker, oci)
    Labels         map[string]string // Image labels
    LayerLabel     []string        // Labels for individual layers
    UnsetLabel     []string        // Labels to remove
    Annotations    map[string]string // OCI annotations
    Manifest       string          // Manifest list name
    IIDFile        string          // File to write image ID
    Identity       string          // Identity label for the image
    OmitHistory    bool           // Omit history in final image

    // Platform Targeting
    Platform      []string        // Target platforms (OS/ARCH[/VARIANT])
    AllPlatforms  bool           // Build for all supported platforms
    Architecture  string         // Override architecture
    OS            string         // Override operating system
    OSFeature     string        // Required OS features
    OSVersion     string        // Override OS version
    Variant       string        // Architecture variant

    // Image Optimization
    Compress           bool     // Compress layers
    DisableCompression bool    // Disable layer compression
    Layers            bool     // Cache intermediate layers
    Squash            bool     // Squash newly built layers
    SquashAll         bool     // Squash all layers

    // Resource Management
    Memory          string     // Memory limit
    MemorySwap      string    // Swap limit
    MemorySwappiness *int64   // Swap usage preference
    CPUPeriod       uint64    // CPU CFS period
    CPUQuota        int64     // CPU CFS quota
    CPUShares       uint64    // CPU shares (relative weight)
    CPUSetCPUs      string    // CPUs in which to allow execution
    CPUSetMems      string    // Memory nodes in which to allow execution
    PidsLimit       *int64    // Process limit

    // Security Configuration
    SecurityOpt      []string  // Security options
    CapAdd          []string  // Add Linux capabilities
    CapDrop         []string  // Drop Linux capabilities
    GroupAdd        []string  // Additional groups
    Isolation       string    // Container isolation technology
    DecryptionKey   string    // Key for encrypted images
    DecryptionKeysPath string // Path to decryption keys
    SignBy          string    // Sign image using key

    // Network Configuration
    Network         string     // Network mode
    DNS            []string   // DNS servers
    DNSSearch      []string   // DNS search domains
    DNSOption      []string   // DNS options
    AddHost        []string   // Additional host-to-IP mappings
    HTTPProxy      bool       // Pass through HTTP proxy
    NetworkAliases []string   // Network-scoped aliases
    NetworkMode    string     // Network mode

    // Namespace Configuration
    CgroupNS       string     // Cgroup namespace mode
    IPC            string     // IPC namespace mode
    PID            string     // PID namespace mode
    UserNS         string     // User namespace mode
    UTS            string     // UTS namespace mode
    IDMappings     string     // Container ID mappings

    // Storage Configuration
    StorageDriver  string     // Storage driver options
    StorageOpt     []string   // Storage driver options
    Volumes        []string   // Bind mount volumes
    ShmSize        string     // Size of /dev/shm

    // Runtime Configuration
    Runtime       string      // Container runtime
    RuntimeFlags  []string    // Runtime flags
    Devices       []string    // Add host devices
    HooksDir      []string    // Hooks directories

    // Authentication
    AuthFile      string      // Path to auth file
    CertDir       string      // Path to certificates
    Credentials   string      // Registry credentials
    TLSVerify     bool       // Require HTTPS and verify certs

    // Process Configuration
    Jobs           uint       // Concurrent jobs
    Stdin         bool       // Keep stdin open
    OOMKillDisable bool      // Disable OOM killer
    OOMScoreAdj    *int      // OOM score adjustment
    Systemd        bool      // Run container init as systemd

    // Build Behavior
    Pull                  string  // Pull policy
    ForceRm              bool    // Always remove intermediate containers
    RemoveIntermediateCtrs bool  // Remove intermediate containers
    Quiet                bool    // Suppress output
    Timestamp            int64   // Set build timestamp
    LogLevel            string   // Logging level
    LogFile             string   // Log file path

    // Host Configuration
    DisableHosts     bool       // Disable host entries
    DisableHostname  bool       // Disable hostname

    // SBOM Options
    SBOMPreset          string  // SBOM generation preset
    SBOMImageOutput     string  // SBOM image output path
    SBOMImagePURLOutput string  // SBOM PURL output path
    SBOMMergeStrategy   string  // SBOM merge strategy
    SBOMOutput          string  // SBOM output path
    SBOMPURLOutput      string  // SBOM PURL output
    SBOMScannerCommand  string  // SBOM scanner command
    SBOMScannerImage    string  // SBOM scanner image

    // Additional Options
    CPPFlags         []string   // C preprocessor flags
    IgnoreFile      string     // Path to ignore file
    Secrets         []string   // Build secrets
    SSHSockets      []string   // SSH agent sockets
    Ulimits         []string   // Ulimit options
    RetryCount      int        // Number of build retries
    RetryDelay      string     // Delay between retries

    // User Namespace Mapping
    UsernsUIDMap     string     // UID map for user namespace
    UsernsGIDMap     string     // GID map for user namespace
    UsernsUIDMapUser string     // Username for UID mapping
    UsernsGIDMapGroup string    // Group name for GID mapping

    // Additional Options
    CgroupParent     string     // Parent cgroup
    IdentityLabel    bool       // Add identity label
}
```

#### Builder
Manages container image building operations:

```go
type Builder struct {
    ssh SSHClientInterface    // SSH client for remote operations
}
```

### Functions

#### NewBuilder
```go
func NewBuilder(ssh SSHClientInterface) *Builder
```
Creates a new Builder instance with the provided SSH client.

#### Build
```go
func (b *Builder) Build(opts BuildOptions) error
```
Builds a container image according to the provided BuildOptions. Constructs and executes a podman build command with all specified options.

#### ValidateContextPath
```go
func (b *Builder) ValidateContextPath(contextPath string) error
```
Validates that the build context path exists and contains a Dockerfile.

#### ListImages
```go
func (b *Builder) ListImages() ([]string, error)
```
Returns a list of all container images in the format "repository:tag".

#### RemoveImage
```go
func (b *Builder) RemoveImage(tag string) error
```
Removes the specified container image.

#### PruneImages
```go
func (b *Builder) PruneImages() error
```
Removes all unused container images to free up space.

#### validateOptions
```go
func (b *Builder) validateOptions(opts BuildOptions) error
```
Validates BuildOptions for correctness, including:
- Platform format validation (OS/ARCH[/VARIANT])
- Pull policy validation (always, missing, never, newer)
- Other option validations as needed

### Usage Example

```go
ssh := NewSSHClient("host", "user", 22, "/path/to/key", "")
builder := NewBuilder(ssh)

opts := BuildOptions{
    File: "Dockerfile",
    ContextPath: "./app",
    Tag: "myapp:latest",
    BuildArgs: map[string]string{
        "VERSION": "1.0.0",
    },
    Platform: []string{"linux/amd64"},
}

if err := builder.Build(opts); err != nil {
    log.Fatalf("Build failed: %v", err)
}
```

## clone.go
```go
type GitCloner struct {
    ssh SSHClientInterface
}

type CloneOptions struct {
    URL         string
    Branch      string
    Destination string
    Depth       int
}

func NewGitCloner(ssh SSHClientInterface) *GitCloner
func (g *GitCloner) Clone(opts CloneOptions) error
func (g *GitCloner) Update(repoPath string, branch string) error
func (g *GitCloner) GetRemoteURL(repoPath string) (string, error)
func (g *GitCloner) IsGitRepository(path string) bool
func (g *GitCloner) GetCurrentCommit(repoPath string) (string, error)
func (g *GitCloner) validateGitInstalled() error
func (g *GitCloner) ensureDirectory(path string) error
func (g *GitCloner) buildCloneCommand(opts CloneOptions) string
```

## container.go
```go
type Healthcheck struct {
    Test     []string
    Interval string
    Timeout  string
    Retries  int
}

type DependencyConfig struct {
    Condition string // "service_started", "service_healthy", "service_completed_successfully"
    Service   string
}

type Container struct {
    Name         string
    Image        string
    Command      []string
    User         string
    Ports        []string
    Volumes      []string
    Environment  map[string]string
    Healthcheck  Healthcheck
    DependsOn    []DependencyConfig
    Deploy       DeployConfig
    // ... many other container configuration options
}

type DeployConfig struct {
    Resources ResourceConfig
    Replicas  *int
    Labels    map[string]string
}

type ResourceConfig struct {
    Limits       Resources
    Reservations Resources
}

type Resources struct {
    CPUs    string
    Memory  string
    Devices []DeviceConfig
}

type DeviceConfig struct {
    Capabilities []string
    Count       *int
    Device      string
    Driver      string
    Options     map[string]string
}

type ContainerManager struct {
    ssh            SSHClientInterface
    volumeManager  VolumeManagerInterface
    networkManager *NetworkManager
}

func NewContainerManager(ssh SSHClientInterface) *ContainerManager
func (c *ContainerManager) Run(container Container) error
func (c *ContainerManager) Remove(name string) error
func (c *ContainerManager) waitForDependencies(container Container) error
func (c *ContainerManager) waitForHealthy(name string) error
func (c *ContainerManager) waitForRunning(name string) error
func (c *ContainerManager) waitForSuccess(name string) error
```

## network.go
```go
type NetworkManager struct {
    ssh SSHClientInterface
}

func NewNetworkManager(ssh SSHClientInterface) *NetworkManager
func (n *NetworkManager) EnsureNetwork(name string, internal bool) error
func (n *NetworkManager) ConnectContainer(container string, networks []string) error
func (n *NetworkManager) DisconnectContainer(container string, network string) error
func (n *NetworkManager) RemoveNetwork(name string) error
func (n *NetworkManager) ListNetworks() ([]string, error)
```

## podman.go
```go
type PodmanInstaller struct {
    ssh SSHClientInterface
}

func NewPodmanInstaller(ssh SSHClientInterface) *PodmanInstaller
func (p *PodmanInstaller) EnsureInstalled() error
func (p *PodmanInstaller) checkAndInstallPodman() error
func (p *PodmanInstaller) checkAndInstallPodmanCompose() error
func (p *PodmanInstaller) detectOS() (string, error)
```

## proxy.go
```go
type ProxyService struct {
    Name        string
    Image       string
    Domains     []string
    SSL         bool
    Networks    []string
    Environment map[string]string
    Volumes     []string
    Command     []string
    Private     bool
    Port        string
    Version     string
    Labels      map[string]string
    Weight      int
    HealthCheckTimeout time.Duration
}

type ProxyManager struct {
    ssh           SSHClientInterface
    proxyBin      string
    containerMgr  *ContainerManager
}

func NewProxyManager(ssh SSHClientInterface, proxyBin string) *ProxyManager
func (p *ProxyManager) DeployService(service ProxyService) error
func (p *ProxyManager) SwitchTraffic(serviceName string, blueVersion, greenVersion string, blueWeight, greenWeight int) error
func (p *ProxyManager) RemoveService(serviceName string) error
func (p *ProxyManager) waitForHealthyContainer(containerName string, timeout time.Duration) error
func (p *ProxyManager) rollbackWeights(serviceName, blueVersion, greenVersion string, blueWeight, greenWeight int) error
```

## registry.go
```go
type RegistryAuth struct {
    Username string
    Password string
}

type ImageRef struct {
    Registry string
    Name     string
    Tag      string
}

type RegistryClient struct {
    ssh       SSHClientInterface
    registry  string
    username  string
    password  string
    authToken string
}

func NewRegistryClient(ssh SSHClientInterface, registry string, auth *RegistryAuth) *RegistryClient
func (r *RegistryClient) PushImage(imageRef ImageRef) error
func (r *RegistryClient) PullImage(imageRef ImageRef) error
func (r *RegistryClient) TagImage(sourceRef, targetRef ImageRef) error
func (r *RegistryClient) ImageExists(imageRef ImageRef) (bool, error)
func (r *RegistryClient) ListTags(imageRef ImageRef) ([]string, error)
func (r *RegistryClient) DeleteImage(imageRef ImageRef) error
func (r *RegistryClient) GetImageDigest(imageRef ImageRef) (string, error)
func (r *RegistryClient) login() error
```

## rexec.go
```go
type RemoteCommand struct {
    client  SSHClientInterface
    command string
    stdout  io.Writer
    stderr  io.Writer
}

func NewRemoteCommand(client SSHClientInterface, command string) *RemoteCommand
func (r *RemoteCommand) SetStdout(w io.Writer)
func (r *RemoteCommand) SetStderr(w io.Writer)
func (r *RemoteCommand) Run() error
func (r *RemoteCommand) Output() ([]byte, error)
func (r *RemoteCommand) CombinedOutput() ([]byte, error)
```

## ssh.go
```go
type SSHClient struct {
    Host     string
    User     string
    Port     int
    client   *ssh.Client
    keyFile  string
    keyData  string
}

type SSHClientInterface interface {
    Connect() error
    Close() error
    GetClient() *ssh.Client
    RunCommand(cmd string) (string, error)
    WriteFile(path string, data []byte, perm uint32) error
    ReadFile(path string) ([]byte, error)
}

func NewSSHClient(host string, user string, port int, keyFile string, keyData string) *SSHClient
func (s *SSHClient) Connect() error
func (s *SSHClient) Close() error
func (s *SSHClient) GetClient() *ssh.Client
func (s *SSHClient) RunCommand(cmd string) (string, error)
func (s *SSHClient) publicKeyAuth() ssh.AuthMethod
```

## volume.go
```go
type VolumeManagerInterface interface {
    EnsureVolume(name string) error
    EnsureVolumes(volumes []string) error
    RemoveVolume(name string) error
    ListVolumes() ([]string, error)
}

type VolumeManager struct {
    ssh SSHClientInterface
}

func NewVolumeManager(ssh SSHClientInterface) *VolumeManager
func (v *VolumeManager) EnsureVolume(name string) error
func (v *VolumeManager) EnsureVolumes(volumes []string) error
func (v *VolumeManager) RemoveVolume(name string) error
func (v *VolumeManager) ListVolumes() ([]string, error)
func (v *VolumeManager) handleSELinux(hostPath string, options []string) error
func (v *VolumeManager) handleVolumeOptions(hostPath string, options []string) []string
```
