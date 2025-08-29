// Package containerengine provides unified container engine detection and management.
package containerengine

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// ContainerEngine represents a container engine that can be used for local cluster management.
type ContainerEngine interface {
	// CheckReady checks if the container engine is available and ready.
	CheckReady() (bool, error)
	// Name returns the name of the container engine.
	Name() string
}

// UnifiedContainerEngine implements the ContainerEngine interface with auto-detection.
type UnifiedContainerEngine struct {
	client client.APIClient
	name   string
}

// NewUnifiedContainerEngine creates a new container engine with auto-detection.
// It tries to connect to a container engine and returns the first available one.
func NewUnifiedContainerEngine() (*UnifiedContainerEngine, error) {
	// Try Docker first (most common)
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err == nil {
		engine := &UnifiedContainerEngine{
			client: dockerClient,
			name:   "Docker",
		}
		if ready, _ := engine.CheckReady(); ready {
			return engine, nil
		}
	}

	// Try Podman with Docker-compatible socket
	podmanClient, err := client.NewClientWithOpts(
		client.WithHost("unix:///run/user/1000/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err == nil {
		engine := &UnifiedContainerEngine{
			client: podmanClient,
			name:   "Podman",
		}
		if ready, _ := engine.CheckReady(); ready {
			return engine, nil
		}
	}

	// Try system-wide Podman socket
	podmanSystemClient, err := client.NewClientWithOpts(
		client.WithHost("unix:///run/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err == nil {
		engine := &UnifiedContainerEngine{
			client: podmanSystemClient,
			name:   "Podman",
		}
		if ready, _ := engine.CheckReady(); ready {
			return engine, nil
		}
	}

	return nil, fmt.Errorf("no container engine (Docker or Podman) available")
}

// CheckReady checks if the container engine is available using the API client.
func (u *UnifiedContainerEngine) CheckReady() (bool, error) {
	ctx := context.Background()

	_, err := u.client.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("%s ping failed: %w", u.name, err)
	}

	return true, nil
}

// Name returns the name of the detected container engine.
func (u *UnifiedContainerEngine) Name() string {
	return u.name
}

// GetClient returns the underlying Docker API client.
func (u *UnifiedContainerEngine) GetClient() client.APIClient {
	return u.client
}