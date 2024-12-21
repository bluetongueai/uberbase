package pkg

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type RegistryClient struct {
	ssh       SSHClientInterface
	registry  string
	username  string
	password  string
	authToken string
}

type RegistryAuth struct {
	Username string
	Password string
}

type ImageRef struct {
	Registry string
	Name     string
	Tag      string
}

func NewRegistryClient(ssh SSHClientInterface, registry string, auth *RegistryAuth) *RegistryClient {
	client := &RegistryClient{
		ssh:      ssh,
		registry: registry,
	}

	if auth != nil {
		client.username = auth.Username
		client.password = auth.Password
		client.authToken = client.createAuthToken(auth)
	}

	return client
}

func (r *RegistryClient) PushImage(imageRef ImageRef) error {
	if err := r.validateImageRef(imageRef); err != nil {
		return fmt.Errorf("invalid image reference: %w", err)
	}
	
	if err := r.login(); err != nil {
		return fmt.Errorf("registry login failed: %w", err)
	}

	name := strings.TrimSuffix(imageRef.Name, ":latest")
	
	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf(
		"podman push %s/%s",
		r.registry,
		name,
	))
	return cmd.Run()
}

func (r *RegistryClient) PullImage(imageRef ImageRef) error {
	if err := r.login(); err != nil {
		return fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf("podman pull %s", fullRef))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to pull image %s: %w", fullRef, err)
	}

	return nil
}

func (r *RegistryClient) TagImage(sourceRef, targetRef ImageRef) error {
	sourceFullRef := r.getFullImageRef(sourceRef)
	targetFullRef := r.getFullImageRef(targetRef)

	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf("podman tag %s %s", sourceFullRef, targetFullRef))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to tag image %s as %s: %w", sourceFullRef, targetFullRef, err)
	}

	return nil
}

func (r *RegistryClient) ImageExists(imageRef ImageRef) (bool, error) {
	if err := r.login(); err != nil {
		return false, fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf("podman inspect %s", fullRef))
	return cmd.Run() == nil, nil
}

func (r *RegistryClient) login() error {
	if r.username == "" || r.password == "" {
		return nil // Skip login if no credentials provided
	}

	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf(
		"podman login -u %s -p %s %s",
		r.username,
		r.password,
		r.registry,
	))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	return nil
}

func (r *RegistryClient) createAuthToken(auth *RegistryAuth) string {
	authJSON, _ := json.Marshal(map[string]string{
		"username": auth.Username,
		"password": auth.Password,
	})
	return base64.StdEncoding.EncodeToString(authJSON)
}

func (r *RegistryClient) getFullImageRef(ref ImageRef) string {
	registry := ref.Registry
	if registry == "" {
		registry = r.registry
	}

	if registry != "" && !strings.HasSuffix(registry, "/") {
		registry += "/"
	}

	tag := ref.Tag
	if tag == "" {
		tag = "latest"
	}

	return fmt.Sprintf("%s%s:%s", registry, ref.Name, tag)
}

func (r *RegistryClient) ListTags(imageRef ImageRef) ([]string, error) {
	if err := r.login(); err != nil {
		return nil, fmt.Errorf("registry login failed: %w", err)
	}

	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf(
		"podman search --list-tags --format '{{.Tag}}' %s/%s",
		r.registry,
		imageRef.Name,
	))

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	return tags, nil
}

func (r *RegistryClient) DeleteImage(imageRef ImageRef) error {
	if err := r.login(); err != nil {
		return fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf("podman rmi %s", fullRef))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to delete image %s: %w", fullRef, err)
	}

	return nil
}

func (r *RegistryClient) GetImageDigest(imageRef ImageRef) (string, error) {
	if err := r.login(); err != nil {
		return "", fmt.Errorf("registry login failed: %w", err)
	}

	fullRef := r.getFullImageRef(imageRef)
	cmd := NewRemoteCommand(r.ssh, fmt.Sprintf(
		"podman inspect --format '{{.Digest}}' %s",
		fullRef,
	))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get image digest: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

func (r *RegistryClient) validateImageRef(ref ImageRef) error {
	if ref.Name == "" {
		return fmt.Errorf("image name is required")
	}
	if ref.Registry != "" {
		// Add registry URL validation
		if strings.Contains(ref.Registry, " ") {
			return fmt.Errorf("invalid registry URL")
		}
	}
	return nil
}
