package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
)

type BuildOptions struct {
	// Basic options
	File         string            // -f, --file pathname or URL
	ContextPath  string
	Tag          string

	// Build arguments and environment
	BuildArgs     map[string]string
	BuildArgFile  string            
	Environment   map[string]string
	Unsetenv     []string

	// Build context and stages
	BuildContexts     map[string]string  
	FromStage        string             
	Target           string             
	SkipUnusedStages bool              

	// Cache options
	NoCache     bool
	CacheFrom   []string
	CacheTo     []string
	CacheTTL    string

	// Image options
	Format         string             
	Labels         map[string]string
	LayerLabel     []string          // --layer-label stringArray
	UnsetLabel     []string
	Annotations    map[string]string
	Manifest       string
	IIDFile        string
	Identity       string
	OmitHistory    bool

	// Platform options
	Platform      []string          // --platform OS/ARCH[/VARIANT]
	AllPlatforms  bool
	Architecture  string
	OS            string
	OSFeature     string
	OSVersion     string
	Variant       string

	// Compression and layers
	Compress           bool
	DisableCompression bool              
	Layers            bool              
	Squash            bool
	SquashAll         bool

	// Resource limits
	Memory          string
	MemorySwap      string
	MemorySwappiness *int64
	CPUPeriod       uint64           // --cpu-period uint
	CPUQuota        int64            // --cpu-quota int
	CPUShares       uint64           // -c, --cpu-shares uint
	CPUSetCPUs      string           // --cpuset-cpus string
	CPUSetMems      string           // --cpuset-mems string
	PidsLimit       *int64
	
	// Security options
	SecurityOpt      []string
	CapAdd          []string
	CapDrop         []string
	GroupAdd        []string         // --group-add strings
	Isolation       string           
	DecryptionKey   string
	DecryptionKeysPath string
	SignBy          string

	// Network options
	Network         string            
	DNS            []string
	DNSSearch      []string
	DNSOption      []string
	AddHost        []string
	HTTPProxy      bool              // --http-proxy
	NetworkAliases []string
	NetworkMode    string         

	// Namespace options
	CgroupNS       string            
	IPC            string            
	PID            string            
	UserNS         string            
	UTS            string            
	IDMappings     string

	// Storage options
	StorageDriver  string
	StorageOpt     []string
	Volumes        []string
	ShmSize        string           

	// Runtime options
	Runtime       string
	RuntimeFlags  []string         // --runtime-flag strings
	Devices       []string
	HooksDir      []string

	// Authentication
	AuthFile      string
	CertDir       string
	Credentials   string
	TLSVerify     bool

	// Process options
	Jobs           uint             
	Stdin         bool
	OOMKillDisable bool
	OOMScoreAdj    *int
	Systemd        bool

	// Build behavior
	Pull                  string             
	ForceRm              bool              
	RemoveIntermediateCtrs bool    
	Quiet                bool
	Timestamp            int64              
	LogLevel            string
	LogFile             string

	// Host configuration
	DisableHosts     bool
	DisableHostname  bool

	// SBOM options
	SBOMPreset          string
	SBOMImageOutput     string
	SBOMImagePURLOutput string
	SBOMMergeStrategy   string
	SBOMOutput          string
	SBOMPURLOutput      string
	SBOMScannerCommand  string
	SBOMScannerImage    string

	// Other options
	CPPFlags         []string          // --cpp-flag stringArray
	IgnoreFile      string
	Secrets         []string
	SSHSockets      []string
	Ulimits         []string
	RetryCount      int              
	RetryDelay      string

	// User namespace mapping options
	UsernsUIDMap     string            
	UsernsGIDMap     string            
	UsernsUIDMapUser string           
	UsernsGIDMapGroup string           

	// Additional options
	CgroupParent     string            
	IdentityLabel    bool              // --identity-label

	// Add these missing fields
	StdinOpen    bool              // --stdin flag
	Creds        string            // --creds string
	NoHosts      bool              // --no-hosts flag
	NoHostname   bool              // --no-hostname flag
	RuntimeFlag  []string          // --runtime-flag strings
}

type Builder struct {
	ssh      SSHClientInterface
}

func NewBuilder(ssh SSHClientInterface) *Builder {
	return &Builder{
		ssh:      ssh,
	}
}

func (b *Builder) Build(opts BuildOptions) error {
	// Validate options first
	if err := b.validateOptions(opts); err != nil {
		return err
	}

	// Validate context path
	if err := b.ValidateContextPath(opts.ContextPath); err != nil {
		return err
	}

	cmd := strings.Builder{}
	cmd.WriteString("podman build")

	// Add file flag if specified
	if opts.File != "" {
		cmd.WriteString(fmt.Sprintf(" -f %s", opts.File))
	}

	// Add tag if specified
	if opts.Tag != "" {
		cmd.WriteString(fmt.Sprintf(" -t %s", opts.Tag))
	}

	// Add new options
	if opts.StdinOpen {
		cmd.WriteString(" --stdin")
	}

	for _, flag := range opts.CPPFlags {
		cmd.WriteString(fmt.Sprintf(" --cpp-flag %s", flag))
	}

	for _, flag := range opts.RuntimeFlags {
		cmd.WriteString(fmt.Sprintf(" --runtime-flag %s", flag))
	}

	if opts.Creds != "" {
		cmd.WriteString(fmt.Sprintf(" --creds %s", opts.Creds))
	}

	// Update renamed options
	if opts.NoHosts {
		cmd.WriteString(" --no-hosts")
	}

	if opts.NoHostname {
		cmd.WriteString(" --no-hostname") 
	}

	// Build arguments and environment
	for key, value := range opts.BuildArgs {
		cmd.WriteString(fmt.Sprintf(" --build-arg %s=%s", key, value))
	}
	if opts.BuildArgFile != "" {
		cmd.WriteString(fmt.Sprintf(" --build-arg-file %s", opts.BuildArgFile))
	}
	for _, env := range opts.Unsetenv {
		cmd.WriteString(fmt.Sprintf(" --unsetenv %s", env))
	}
	for key, value := range opts.Environment {
		cmd.WriteString(fmt.Sprintf(" --env %s=%s", key, value))
	}

	// Build context and stages
	for k, v := range opts.BuildContexts {
		cmd.WriteString(fmt.Sprintf(" --build-context %s=%s", k, v))
	}
	if opts.FromStage != "" {
		cmd.WriteString(fmt.Sprintf(" --from %s", opts.FromStage))
	}
	if opts.Target != "" {
		cmd.WriteString(fmt.Sprintf(" --target %s", opts.Target))
	}
	if opts.SkipUnusedStages {
		cmd.WriteString(" --skip-unused-stages")
	}

	// Cache options
	if opts.NoCache {
		cmd.WriteString(" --no-cache")
	}
	for _, cacheFrom := range opts.CacheFrom {
		cmd.WriteString(fmt.Sprintf(" --cache-from %s", cacheFrom))
	}
	for _, cacheTo := range opts.CacheTo {
		cmd.WriteString(fmt.Sprintf(" --cache-to %s", cacheTo))
	}
	if opts.CacheTTL != "" {
		cmd.WriteString(fmt.Sprintf(" --cache-ttl %s", opts.CacheTTL))
	}

	// Image options
	if opts.Format != "" {
		cmd.WriteString(fmt.Sprintf(" --format %s", opts.Format))
	}
	for k, v := range opts.Labels {
		cmd.WriteString(fmt.Sprintf(" --label %s=%s", k, v))
	}
	for _, label := range opts.LayerLabel {
		cmd.WriteString(fmt.Sprintf(" --layer-label %s", label))
	}
	for _, label := range opts.UnsetLabel {
		cmd.WriteString(fmt.Sprintf(" --unsetlabel %s", label))
	}
	for k, v := range opts.Annotations {
		cmd.WriteString(fmt.Sprintf(" --annotation %s=%s", k, v))
	}
	if opts.Manifest != "" {
		cmd.WriteString(fmt.Sprintf(" --manifest %s", opts.Manifest))
	}
	if opts.IIDFile != "" {
		cmd.WriteString(fmt.Sprintf(" --iidfile %s", opts.IIDFile))
	}
	if opts.Identity != "" {
		cmd.WriteString(fmt.Sprintf(" --identity %s", opts.Identity))
	}
	if opts.OmitHistory {
		cmd.WriteString(" --omit-history")
	}

	// Platform options
	if len(opts.Platform) > 0 {
		for _, platform := range opts.Platform {
			cmd.WriteString(fmt.Sprintf(" --platform %s", platform))
		}
	}
	if opts.AllPlatforms {
		cmd.WriteString(" --all-platforms")
	}
	if opts.Architecture != "" {
		cmd.WriteString(fmt.Sprintf(" --arch %s", opts.Architecture))
	}
	if opts.OS != "" {
		cmd.WriteString(fmt.Sprintf(" --os %s", opts.OS))
	}
	if opts.OSFeature != "" {
		cmd.WriteString(fmt.Sprintf(" --os-feature %s", opts.OSFeature))
	}
	if opts.OSVersion != "" {
		cmd.WriteString(fmt.Sprintf(" --os-version %s", opts.OSVersion))
	}
	if opts.Variant != "" {
		cmd.WriteString(fmt.Sprintf(" --variant %s", opts.Variant))
	}

	// Compression and layers
	if opts.Compress {
		cmd.WriteString(" --compress")
	}
	if opts.DisableCompression {
		cmd.WriteString(" --disable-compression")
	}
	if opts.Layers {
		cmd.WriteString(" --layers")
	}
	if opts.Squash {
		cmd.WriteString(" --squash")
	}
	if opts.SquashAll {
		cmd.WriteString(" --squash-all")
	}

	// Resource limits
	if opts.Memory != "" {
		cmd.WriteString(fmt.Sprintf(" --memory %s", opts.Memory))
	}
	if opts.MemorySwap != "" {
		cmd.WriteString(fmt.Sprintf(" --memory-swap %s", opts.MemorySwap))
	}
	if opts.MemorySwappiness != nil {
		cmd.WriteString(fmt.Sprintf(" --memory-swappiness %d", *opts.MemorySwappiness))
	}
	if opts.CPUPeriod != 0 {
		cmd.WriteString(fmt.Sprintf(" --cpu-period %d", opts.CPUPeriod))
	}
	if opts.CPUQuota != 0 {
		cmd.WriteString(fmt.Sprintf(" --cpu-quota %d", opts.CPUQuota))
	}
	if opts.CPUShares != 0 {
		cmd.WriteString(fmt.Sprintf(" --cpu-shares %d", opts.CPUShares))
	}
	if opts.CPUSetCPUs != "" {
		cmd.WriteString(fmt.Sprintf(" --cpuset-cpus %s", opts.CPUSetCPUs))
	}
	if opts.CPUSetMems != "" {
		cmd.WriteString(fmt.Sprintf(" --cpuset-mems %s", opts.CPUSetMems))
	}
	if opts.PidsLimit != nil {
		cmd.WriteString(fmt.Sprintf(" --pids-limit %d", *opts.PidsLimit))
	}

	// Security options
	for _, opt := range opts.SecurityOpt {
		cmd.WriteString(fmt.Sprintf(" --security-opt %s", opt))
	}
	for _, cap := range opts.CapAdd {
		cmd.WriteString(fmt.Sprintf(" --cap-add %s", cap))
	}
	for _, cap := range opts.CapDrop {
		cmd.WriteString(fmt.Sprintf(" --cap-drop %s", cap))
	}
	for _, group := range opts.GroupAdd {
		cmd.WriteString(fmt.Sprintf(" --group-add %s", group))
	}
	if opts.Isolation != "" {
		cmd.WriteString(fmt.Sprintf(" --isolation %s", opts.Isolation))
	}
	if opts.DecryptionKey != "" {
		cmd.WriteString(fmt.Sprintf(" --decryption-key %s", opts.DecryptionKey))
	}
	if opts.DecryptionKeysPath != "" {
		cmd.WriteString(fmt.Sprintf(" --decryption-keys-path %s", opts.DecryptionKeysPath))
	}
	if opts.SignBy != "" {
		cmd.WriteString(fmt.Sprintf(" --sign-by %s", opts.SignBy))
	}

	// Network options
	if opts.Network != "" {
		cmd.WriteString(fmt.Sprintf(" --network %s", opts.Network))
	}
	for _, dns := range opts.DNS {
		cmd.WriteString(fmt.Sprintf(" --dns %s", dns))
	}
	for _, search := range opts.DNSSearch {
		cmd.WriteString(fmt.Sprintf(" --dns-search %s", search))
	}
	for _, option := range opts.DNSOption {
		cmd.WriteString(fmt.Sprintf(" --dns-option %s", option))
	}
	for _, host := range opts.AddHost {
		cmd.WriteString(fmt.Sprintf(" --add-host %s", host))
	}
	if opts.HTTPProxy {
		cmd.WriteString(" --http-proxy")
	}

	// Namespace options
	if opts.CgroupNS != "" {
		cmd.WriteString(fmt.Sprintf(" --cgroupns %s", opts.CgroupNS))
	}
	if opts.IPC != "" {
		cmd.WriteString(fmt.Sprintf(" --ipc %s", opts.IPC))
	}
	if opts.PID != "" {
		cmd.WriteString(fmt.Sprintf(" --pid %s", opts.PID))
	}
	if opts.UserNS != "" {
		cmd.WriteString(fmt.Sprintf(" --userns %s", opts.UserNS))
	}
	if opts.UTS != "" {
		cmd.WriteString(fmt.Sprintf(" --uts %s", opts.UTS))
	}
	if opts.IDMappings != "" {
		cmd.WriteString(fmt.Sprintf(" --id-mappings %s", opts.IDMappings))
	}

	// Storage options
	if opts.StorageDriver != "" {
		cmd.WriteString(fmt.Sprintf(" --storage-driver %s", opts.StorageDriver))
	}
	for _, opt := range opts.StorageOpt {
		cmd.WriteString(fmt.Sprintf(" --storage-opt %s", opt))
	}
	for _, volume := range opts.Volumes {
		cmd.WriteString(fmt.Sprintf(" --volume %s", volume))
	}
	if opts.ShmSize != "" {
		cmd.WriteString(fmt.Sprintf(" --shm-size %s", opts.ShmSize))
	}

	// Runtime options
	if opts.Runtime != "" {
		cmd.WriteString(fmt.Sprintf(" --runtime %s", opts.Runtime))
	}
	for _, flag := range opts.RuntimeFlags {
		cmd.WriteString(fmt.Sprintf(" --runtime-flag %s", flag))
	}
	for _, device := range opts.Devices {
		cmd.WriteString(fmt.Sprintf(" --device %s", device))
	}
	for _, hook := range opts.HooksDir {
		cmd.WriteString(fmt.Sprintf(" --hooks-dir %s", hook))
	}

	// Authentication
	if opts.AuthFile != "" {
		cmd.WriteString(fmt.Sprintf(" --authfile %s", opts.AuthFile))
	}
	if opts.CertDir != "" {
		cmd.WriteString(fmt.Sprintf(" --cert-dir %s", opts.CertDir))
	}
	if opts.Credentials != "" {
		cmd.WriteString(fmt.Sprintf(" --credentials %s", opts.Credentials))
	}
	if opts.TLSVerify {
		cmd.WriteString(" --tls-verify")
	}

	// Process options
	if opts.Jobs != 0 {
		cmd.WriteString(fmt.Sprintf(" --jobs %d", opts.Jobs))
	}
	if opts.Stdin {
		cmd.WriteString(" --stdin")
	}
	if opts.OOMKillDisable {
		cmd.WriteString(" --oom-kill-disable")
	}
	if opts.OOMScoreAdj != nil {
		cmd.WriteString(fmt.Sprintf(" --oom-score-adj %d", *opts.OOMScoreAdj))
	}
	if opts.Systemd {
		cmd.WriteString(" --systemd")
	}

	// Build behavior
	if opts.Pull != "" {
		cmd.WriteString(fmt.Sprintf(" --pull %s", opts.Pull))
	}
	if opts.ForceRm {
		cmd.WriteString(" --force-rm")
	}
	if opts.RemoveIntermediateCtrs {
		cmd.WriteString(" --rm")
	}
	if opts.Quiet {
		cmd.WriteString(" --quiet")
	}
	if opts.Timestamp != 0 {
		cmd.WriteString(fmt.Sprintf(" --timestamp %d", opts.Timestamp))
	}
	if opts.LogLevel != "" {
		cmd.WriteString(fmt.Sprintf(" --log-level %s", opts.LogLevel))
	}
	if opts.LogFile != "" {
		cmd.WriteString(fmt.Sprintf(" --log-file %s", opts.LogFile))
	}

	// Host configuration
	if opts.DisableHosts {
		cmd.WriteString(" --disable-hosts")
	}
	if opts.DisableHostname {
		cmd.WriteString(" --disable-hostname")
	}

	// SBOM options
	if opts.SBOMPreset != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-preset %s", opts.SBOMPreset))
	}
	if opts.SBOMImageOutput != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-image-output %s", opts.SBOMImageOutput))
	}
	if opts.SBOMImagePURLOutput != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-image-purl-output %s", opts.SBOMImagePURLOutput))
	}
	if opts.SBOMMergeStrategy != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-merge-strategy %s", opts.SBOMMergeStrategy))
	}
	if opts.SBOMOutput != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-output %s", opts.SBOMOutput))
	}
	if opts.SBOMPURLOutput != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-purl-output %s", opts.SBOMPURLOutput))
	}
	if opts.SBOMScannerCommand != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-scanner-command %s", opts.SBOMScannerCommand))
	}
	if opts.SBOMScannerImage != "" {
		cmd.WriteString(fmt.Sprintf(" --sbom-scanner-image %s", opts.SBOMScannerImage))
	}

	// Other options
	if opts.IgnoreFile != "" {
		cmd.WriteString(fmt.Sprintf(" --ignorefile %s", opts.IgnoreFile))
	}
	for _, secret := range opts.Secrets {
		cmd.WriteString(fmt.Sprintf(" --secret %s", secret))
	}
	for _, socket := range opts.SSHSockets {
		cmd.WriteString(fmt.Sprintf(" --ssh %s", socket))
	}
	for _, limit := range opts.Ulimits {
		cmd.WriteString(fmt.Sprintf(" --ulimit %s", limit))
	}
	if opts.RetryCount != 0 {
		cmd.WriteString(fmt.Sprintf(" --retry %d", opts.RetryCount))
	}
	if opts.RetryDelay != "" {
		cmd.WriteString(fmt.Sprintf(" --retry-delay %s", opts.RetryDelay))
	}

	// Add context path at the end
	cmd.WriteString(" " + opts.ContextPath)

	// Add new user namespace mapping options
	if opts.UsernsUIDMap != "" {
		cmd.WriteString(fmt.Sprintf(" --userns-uid-map %s", opts.UsernsUIDMap))
	}
	if opts.UsernsGIDMap != "" {
		cmd.WriteString(fmt.Sprintf(" --userns-gid-map %s", opts.UsernsGIDMap))
	}
	if opts.UsernsUIDMapUser != "" {
		cmd.WriteString(fmt.Sprintf(" --userns-uid-map-user %s", opts.UsernsUIDMapUser))
	}
	if opts.UsernsGIDMapGroup != "" {
		cmd.WriteString(fmt.Sprintf(" --userns-gid-map-group %s", opts.UsernsGIDMapGroup))
	}

	// Add missing options
	if opts.CgroupParent != "" {
		cmd.WriteString(fmt.Sprintf(" --cgroup-parent %s", opts.CgroupParent))
	}

	// Add handling for new options
	if opts.IdentityLabel {
		cmd.WriteString(" --identity-label")
	}
	for _, flag := range opts.CPPFlags {
		cmd.WriteString(fmt.Sprintf(" --cpp-flag %s", flag))
	}

	// Update command building to use StdinOpen
	if opts.StdinOpen {
		cmd.WriteString(" --stdin")
	}

	// Add RuntimeFlags support
	for _, flag := range opts.RuntimeFlags {
		cmd.WriteString(fmt.Sprintf(" --runtime-flag %s", flag))
	}

	// Execute the command
	execCmd := NewRemoteCommand(b.ssh, cmd.String())
	err := execCmd.Run()
	
	// Handle cleanup on failure if ForceRm is true
	if err != nil {
		if opts.ForceRm {
			cleanupCmd := NewRemoteCommand(b.ssh, "podman rm $(podman ps -a -q -f status=exited)")
			_ = cleanupCmd.Run() // Ignore cleanup errors
		}
		return fmt.Errorf("build failed: %w", err)
	}
	
	return nil
}

func (b *Builder) ValidateContextPath(contextPath string) error {
	if contextPath == "" {
		return fmt.Errorf("context path cannot be empty")
	}

	// Check if context path exists
	cmd := NewRemoteCommand(b.ssh, fmt.Sprintf("test -d %s", contextPath))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("context path does not exist or is not accessible: %s", contextPath)
	}

	// Check if Dockerfile exists in context
	dockerfile := filepath.Join(contextPath, "Dockerfile")
	cmd = NewRemoteCommand(b.ssh, fmt.Sprintf("test -f %s", dockerfile))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("dockerfile not found in context: %s", dockerfile)
	}

	return nil
}

func (b *Builder) ListImages() ([]string, error) {
	cmd := NewRemoteCommand(b.ssh, "podman images --format '{{.Repository}}:{{.Tag}}'")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	images := strings.Split(strings.TrimSpace(string(output)), "\n")
	return images, nil
}

func (b *Builder) RemoveImage(tag string) error {
	cmd := NewRemoteCommand(b.ssh, fmt.Sprintf("podman rmi %s", tag))
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to remove image %s: %w", tag, err)
	}
	if strings.Contains(string(output), "Error") {
		return fmt.Errorf("failed to remove image %s: %s", tag, string(output))
	}
	return nil
}

func (b *Builder) PruneImages() error {
	cmd := NewRemoteCommand(b.ssh, "podman image prune -f")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to prune images: %w", err)
	}
	if strings.Contains(string(output), "Error") {
		return fmt.Errorf("failed to prune images: %s", string(output))
	}
	return nil
}

func (b *Builder) validateOptions(opts BuildOptions) error {
	var errors []string

	// Validate required fields first
	if opts.ContextPath == "" {
		errors = append(errors, "context path is required")
	}
	
	if opts.Tag == "" {
		errors = append(errors, "tag is required")
	}

	// Validate platform format
	for _, platform := range opts.Platform {
		if platform != "" && !strings.Contains(platform, "/") {
			errors = append(errors, fmt.Sprintf("invalid platform format %s - must be OS/ARCH[/VARIANT]", platform))
		}
	}

	// Validate pull values
	validPull := map[string]bool{
		"always": true, "true": true,
		"missing": true,
		"never": true, "false": true,
		"newer": true,
	}
	if opts.Pull != "" && !validPull[opts.Pull] {
		errors = append(errors, fmt.Sprintf("invalid pull value: %s", opts.Pull))
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
