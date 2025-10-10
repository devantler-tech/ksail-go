// Package registry provides functionality for managing Docker registry containers
// used as mirror registries for Kind and K3d clusters.
package registry

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const (
	// DefaultRegistryImage is the default Docker registry image to use.
	DefaultRegistryImage = "registry:3"
)

// RegistryConfig describes a registry container configuration.
type RegistryConfig struct {
	Name     string // Container name
	HostPort string // Host port to bind (e.g., "5000")
	Image    string // Registry image (default: registry:3)
}

// Manager handles Docker registry container lifecycle.
type Manager struct {
	client client.APIClient
}

// NewManager creates a new registry manager with the provided Docker client.
func NewManager(dockerClient client.APIClient) *Manager {
	return &Manager{
		client: dockerClient,
	}
}

// CreateRegistry creates a Docker registry container if it doesn't already exist.
// Returns nil if the registry already exists or was successfully created.
func (m *Manager) CreateRegistry(ctx context.Context, cfg RegistryConfig) error {
	if cfg.Image == "" {
		cfg.Image = DefaultRegistryImage
	}

	// Check if container already exists
	exists, err := m.containerExists(ctx, cfg.Name)
	if err != nil {
		return fmt.Errorf("failed to check if registry exists: %w", err)
	}

	if exists {
		return nil // Registry already exists
	}

	// Pull the registry image
	err = m.pullImage(ctx, cfg.Image)
	if err != nil {
		return fmt.Errorf("failed to pull registry image: %w", err)
	}

	// Create container configuration
	containerPort := nat.Port("5000/tcp")
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: cfg.HostPort,
	}

	containerConfig := &container.Config{
		Image: cfg.Image,
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{hostBinding},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}

	networkConfig := &network.NetworkingConfig{}

	// Create the container
	resp, err := m.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		cfg.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create registry container: %w", err)
	}

	// Start the container
	err = m.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}

	return nil
}

// containerExists checks if a container with the given name exists.
func (m *Manager) containerExists(ctx context.Context, name string) (bool, error) {
	containers, err := m.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, n := range c.Names {
			// Docker container names start with "/"
			cleanName := strings.TrimPrefix(n, "/")
			if cleanName == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// pullImage pulls a Docker image if not already present.
func (m *Manager) pullImage(ctx context.Context, imageName string) error {
	reader, err := m.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()

	// Consume the reader to ensure the pull completes
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	return nil
}
