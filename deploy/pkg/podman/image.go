package podman

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

// BuildOptions represents configuration for building an image
type BuildOptions struct {
	// Core build options
	File        string            // Dockerfile path
	ContextPath string            // Build context path
	Tag         ImageRef          // Image tag
	BuildArgs   map[string]string // Build arguments

	// Basic build behavior
	ForceRm bool   // Remove intermediate containers
	NoCache bool   // Don't use cache
	Pull    string // Pull policy (always, missing, never)
	Quiet   bool   // Suppress output

	// Authentication
	AuthFile string // Path to authentication file
}

// Builder manages image building operations
type Builder struct {
	ssh *core.SSHConnection
}

// NewBuilder creates a new image builder
func NewBuilder(ssh *core.SSHConnection) *Builder {
	core.Logger.Debug("Creating new image Builder")
	return &Builder{
		ssh: ssh,
	}
}

// Build builds a container image using the provided options
func (b *Builder) Build(opts BuildOptions) error {
	if err := b.validateBuildOptions(opts); err != nil {
		return fmt.Errorf("invalid build options: %w", err)
	}

	cmd := strings.Builder{}
	cmd.WriteString("podman build")

	// Core options
	if opts.File != "" {
		cmd.WriteString(fmt.Sprintf(" -f %s", opts.File))
	}
	if opts.Tag.String() != "" {
		cmd.WriteString(fmt.Sprintf(" -t %s", opts.Tag.String()))
	}
	for k, v := range opts.BuildArgs {
		cmd.WriteString(fmt.Sprintf(" --build-arg %s=%s", k, v))
	}

	// Build behavior
	if opts.ForceRm {
		cmd.WriteString(" --force-rm")
	}
	if opts.NoCache {
		cmd.WriteString(" --no-cache")
	}
	if opts.Pull != "" {
		cmd.WriteString(fmt.Sprintf(" --pull=%s", opts.Pull))
	}
	if opts.Quiet {
		cmd.WriteString(" --quiet")
	}

	// Authentication
	if opts.AuthFile != "" {
		cmd.WriteString(fmt.Sprintf(" --authfile %s", opts.AuthFile))
	}

	// Add context path at the end
	cmd.WriteString(" " + opts.ContextPath)

	// Execute the build
	_, err := b.ssh.Exec(cmd.String())
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

// Tag applies a new tag to an existing image
func (b *Builder) Tag(source, target string) error {
	core.Logger.Infof("Tagging image %s as %s", source, target)
	cmd := fmt.Sprintf("podman tag %s %s", source, target)
	if _, err := b.ssh.Exec(cmd); err != nil {
		core.Logger.Errorf("Failed to tag image: %v", err)
		return fmt.Errorf("failed to tag image: %w", err)
	}
	return nil
}

// Remove deletes an image
func (b *Builder) Remove(tag string, force bool) error {
	core.Logger.Infof("Removing image: %s", tag)
	cmd := fmt.Sprintf("podman rmi %s", tag)
	if force {
		cmd += " -f"
	}
	if _, err := b.ssh.Exec(cmd); err != nil {
		core.Logger.Errorf("Failed to remove image: %v", err)
		return fmt.Errorf("failed to remove image: %w", err)
	}
	return nil
}

// Pull downloads an image from a registry
func (b *Builder) Pull(image string) error {
	core.Logger.Infof("Pulling image: %s", image)
	cmd := fmt.Sprintf("podman pull %s", image)
	if _, err := b.ssh.Exec(cmd); err != nil {
		core.Logger.Errorf("Failed to pull image: %v", err)
		return fmt.Errorf("failed to pull image: %w", err)
	}
	return nil
}

// List returns a list of all images
func (b *Builder) List() ([]string, error) {
	core.Logger.Debug("Listing images")
	output, err := b.ssh.Exec("podman images --format '{{.Repository}}:{{.Tag}}'")
	if err != nil {
		core.Logger.Errorf("Failed to list images: %v", err)
		return nil, fmt.Errorf("failed to list images: %w", err)
	}

	images := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(images) == 1 && images[0] == "" {
		return []string{}, nil
	}
	return images, nil
}

// Exists checks if an image exists locally
func (b *Builder) Exists(tag string) (bool, error) {
	core.Logger.Debugf("Checking if image exists: %s", tag)
	cmd := fmt.Sprintf("podman image exists %s", tag)
	_, err := b.ssh.Exec(cmd)
	if err != nil {
		if strings.Contains(err.Error(), "no such image") {
			return false, nil
		}
		core.Logger.Errorf("Error checking image existence: %v", err)
		return false, fmt.Errorf("error checking image existence: %w", err)
	}
	return true, nil
}

func (b *Builder) validateBuildOptions(opts BuildOptions) error {
	var errors []string

	if opts.ContextPath == "" {
		errors = append(errors, "context path is required")
	}
	if opts.Tag.Name == "" {
		errors = append(errors, "tag is required")
	}

	validPull := map[string]bool{
		"always":  true,
		"missing": true,
		"never":   true,
	}
	if opts.Pull != "" && !validPull[opts.Pull] {
		errors = append(errors, fmt.Sprintf("invalid pull value: %s", opts.Pull))
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

type ImageManager struct {
	ssh      *core.SSHConnection
	registry *RegistryClient
}

func NewImageManager(ssh *core.SSHConnection, registryConfig RegistryConfig) *ImageManager {
	return &ImageManager{
		ssh:      ssh,
		registry: New(ssh, registryConfig),
	}
}

func (i *ImageManager) Build(opts BuildOptions) error {
	if err := i.validateBuildOptions(opts); err != nil {
		return fmt.Errorf("invalid build options: %w", err)
	}

	cmd := strings.Builder{}
	cmd.WriteString("podman build")

	// Core options
	if opts.File != "" {
		cmd.WriteString(fmt.Sprintf(" -f %s", opts.File))
	}
	if opts.Tag.String() != "" {
		cmd.WriteString(fmt.Sprintf(" -t %s", opts.Tag.String()))
	}
	for k, v := range opts.BuildArgs {
		cmd.WriteString(fmt.Sprintf(" --build-arg %s=%s", k, v))
	}

	// Build behavior
	if opts.ForceRm {
		cmd.WriteString(" --force-rm")
	}
	if opts.NoCache {
		cmd.WriteString(" --no-cache")
	}
	if opts.Pull != "" {
		cmd.WriteString(fmt.Sprintf(" --pull=%s", opts.Pull))
	}
	if opts.Quiet {
		cmd.WriteString(" --quiet")
	}

	// Authentication
	if opts.AuthFile != "" {
		cmd.WriteString(fmt.Sprintf(" --authfile %s", opts.AuthFile))
	}

	// Add context path at the end
	cmd.WriteString(" " + opts.ContextPath)

	// Execute the build
	_, err := i.ssh.Exec(cmd.String())
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// If registry is configured, push the image
	if i.registry != nil && i.registry.IsConfigured() {
		if err := i.registry.PushImage(opts.Tag); err != nil {
			return fmt.Errorf("push after build failed: %w", err)
		}
	}

	return nil
}

func (i *ImageManager) EnsureImage(ref ImageRef) error {
	// Check if image exists locally
	exists, err := i.registry.ImageExists(ref)
	if err != nil {
		return err
	}

	if !exists {
		// Try to pull the image
		if err := i.registry.PullImage(ref); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	}

	return nil
}

func (i *ImageManager) RemoveImage(ref ImageRef) error {
	return i.registry.DeleteImage(ref)
}

func (i *ImageManager) ListImages() ([]ImageRef, error) {
	output, err := i.ssh.Exec("podman images --format '{{.Repository}}:{{.Tag}}'")
	if err != nil {
		return nil, err
	}

	var refs []ImageRef
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			refs = append(refs, ParseImageRef(line))
		}
	}
	return refs, nil
}

func (i *ImageManager) validateBuildOptions(opts BuildOptions) error {
	var errors []string

	if opts.ContextPath == "" {
		errors = append(errors, "context path is required")
	}
	if opts.Tag.Name == "" {
		errors = append(errors, "tag is required")
	}

	validPull := map[string]bool{
		"always":  true,
		"missing": true,
		"never":   true,
	}
	if opts.Pull != "" && !validPull[opts.Pull] {
		errors = append(errors, fmt.Sprintf("invalid pull value: %s", opts.Pull))
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}
