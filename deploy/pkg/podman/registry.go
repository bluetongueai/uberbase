package podman

import (
	"fmt"
	"strings"

	"github.com/bluetongueai/uberbase/deploy/pkg/core"
)

var logger = core.Logger

type ImageRef struct {
	Registry string // Optional, defaults to docker.io
	Name     string
	Tag      string
}

// ParseImageRef parses a string into an ImageRef
// Examples:
// - "nginx" -> {Registry: "docker.io", Name: "nginx", Tag: "latest"}
// - "nginx:1.19" -> {Registry: "docker.io", Name: "nginx", Tag: "1.19"}
// - "my-registry.com/app:v1" -> {Registry: "my-registry.com", Name: "app", Tag: "v1"}
func ParseImageRef(image string) ImageRef {
	ref := ImageRef{}

	// Split registry and rest
	parts := strings.Split(image, "/")
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		ref.Registry = parts[0]
		image = strings.Join(parts[1:], "/")
	} else {
		ref.Registry = "docker.io"
	}

	// Split name and tag
	parts = strings.Split(image, ":")
	ref.Name = parts[0]
	if len(parts) > 1 {
		ref.Tag = parts[1]
	} else {
		ref.Tag = "latest"
	}

	return ref
}

func (r ImageRef) String() string {
	if r.Registry == "docker.io" {
		if r.Tag == "latest" {
			return r.Name
		}
		return fmt.Sprintf("%s:%s", r.Name, r.Tag)
	}
	return fmt.Sprintf("%s/%s:%s", r.Registry, r.Name, r.Tag)
}

type RegistryConfig struct {
	Host     string
	Username string
	Password string
}

type RegistryClient struct {
	host     string
	username string
	password string
	ssh      *core.SSHConnection
}

func New(ssh *core.SSHConnection, config RegistryConfig) *RegistryClient {
	return &RegistryClient{
		host:     config.Host,
		username: config.Username,
		password: config.Password,
		ssh:      ssh,
	}
}

func (r *RegistryClient) PushImage(imageRef ImageRef) error {
	logger.Infof("Pushing image: %s", imageRef.Name)
	if err := r.validateImageRef(imageRef); err != nil {
		logger.Errorf("Invalid image reference: %v", err)
		return fmt.Errorf("invalid image reference: %w", err)
	}

	if err := r.login(); err != nil {
		logger.Errorf("Registry login failed: %v", err)
		return fmt.Errorf("registry login failed: %w", err)
	}

	name := strings.TrimSuffix(imageRef.Name, ":latest")

	_, err := r.ssh.Exec(fmt.Sprintf(
		"podman push %s/%s",
		r.host,
		name,
	))
	if err != nil {
		logger.Errorf("Failed to push image: %v", err)
		return err
	}
	return nil
}

func (r *RegistryClient) PullImage(imageRef ImageRef) error {
	logger.Infof("Pulling image: %s", imageRef.Name)
	if err := r.login(); err != nil {
		logger.Errorf("Registry login failed: %v", err)
		return fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	_, err := r.ssh.Exec(fmt.Sprintf("podman pull %s", fullRef))
	if err != nil {
		logger.Errorf("Failed to pull image: %v", err)
		return fmt.Errorf("failed to pull image %s: %w", fullRef, err)
	}

	return nil
}

func (r *RegistryClient) TagImage(sourceRef, targetRef ImageRef) error {
	logger.Infof("Tagging image from %s to %s", sourceRef.Name, targetRef.Name)
	sourceFullRef := r.getFullImageRef(sourceRef)
	targetFullRef := r.getFullImageRef(targetRef)

	_, err := r.ssh.Exec(fmt.Sprintf("podman tag %s %s", sourceFullRef, targetFullRef))
	if err != nil {
		logger.Errorf("Failed to tag image: %v", err)
		return fmt.Errorf("failed to tag image %s as %s: %w", sourceFullRef, targetFullRef, err)
	}

	return nil
}

func (r *RegistryClient) ImageExists(imageRef ImageRef) (bool, error) {
	logger.Infof("Checking if image exists: %s", imageRef.Name)
	if err := r.login(); err != nil {
		logger.Errorf("Registry login failed: %v", err)
		return false, fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	_, err := r.ssh.Exec(fmt.Sprintf("podman inspect %s", fullRef))
	exists := err == nil
	if !exists {
		logger.Warnf("Image does not exist: %s", imageRef.Name)
	}
	return exists, nil
}

func (r *RegistryClient) ListTags(imageRef ImageRef) ([]string, error) {
	logger.Infof("Listing tags for image: %s", imageRef.Name)
	if err := r.login(); err != nil {
		logger.Errorf("Registry login failed: %v", err)
		return nil, fmt.Errorf("registry login failed: %w", err)
	}

	output, err := r.ssh.Exec(fmt.Sprintf(
		"podman search --list-tags --format '{{.Tag}}' %s/%s",
		r.host,
		imageRef.Name,
	))
	if err != nil {
		logger.Errorf("Failed to list tags: %v", err)
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	return tags, nil
}

func (r *RegistryClient) DeleteImage(imageRef ImageRef) error {
	logger.Infof("Deleting image: %s", imageRef.Name)
	if err := r.login(); err != nil {
		logger.Errorf("Registry login failed: %v", err)
		return fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	_, err := r.ssh.Exec(fmt.Sprintf("podman rmi %s", fullRef))
	if err != nil {
		logger.Errorf("Failed to delete image: %v", err)
		return fmt.Errorf("failed to delete image %s: %w", fullRef, err)
	}

	return nil
}

func (r *RegistryClient) GetImageDigest(imageRef ImageRef) (string, error) {
	logger.Infof("Getting image digest for: %s", imageRef.Name)
	if err := r.login(); err != nil {
		logger.Errorf("Registry login failed: %v", err)
		return "", fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	output, err := r.ssh.Exec(fmt.Sprintf(
		"podman inspect --format '{{.Digest}}' %s",
		fullRef,
	))
	if err != nil {
		logger.Errorf("Failed to get image digest: %v", err)
		return "", fmt.Errorf("failed to get image digest: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *RegistryClient) validateImageRef(ref ImageRef) error {
	if ref.Name == "" {
		return fmt.Errorf("image name is required")
	}
	if r.host != "" {
		// Add registry URL validation
		if strings.Contains(r.host, " ") {
			return fmt.Errorf("invalid registry URL")
		}
	}
	return nil
}

func (r *RegistryClient) login() error {
	if r.username == "" || r.password == "" {
		logger.Warn("No credentials provided, skipping login")
		return nil // Skip login if no credentials provided
	}

	_, err := r.ssh.Exec(fmt.Sprintf(
		"podman login -u %s -p %s %s",
		r.username,
		r.password,
		r.host,
	))
	if err != nil {
		logger.Errorf("Login failed: %v", err)
		return fmt.Errorf("login failed: %w", err)
	}

	logger.Info("Registry login successful")
	return nil
}

func (r *RegistryClient) getFullImageRef(ref ImageRef) string {
	registry := r.host

	if registry != "" && !strings.HasSuffix(registry, "/") {
		registry += "/"
	}

	tag := ref.Tag
	if tag == "" {
		tag = "latest"
	}

	return fmt.Sprintf("%s%s:%s", registry, ref.Name, tag)
}

func (r *RegistryClient) IsConfigured() bool {
	return r.host != ""
}
