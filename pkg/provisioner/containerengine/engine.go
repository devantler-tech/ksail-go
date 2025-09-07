// Package containerengine provides unified container engine detection and management.
package containerengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/client"
)

// ErrNoContainerEngine is returned when no container engine (Docker or Podman) is available.
var ErrNoContainerEngine = errors.New("no container engine (Docker or Podman) available")

// ClientFactory defines how to create container engine clients
type ClientFactory interface {
	NewDockerClient() (client.APIClient, error)
	NewPodmanUserClient() (client.APIClient, error)
	NewPodmanSystemClient() (client.APIClient, error)
}

// DefaultClientFactory implements ClientFactory using the actual Docker client
type DefaultClientFactory struct{}

func (f *DefaultClientFactory) NewDockerClient() (client.APIClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

func (f *DefaultClientFactory) NewPodmanUserClient() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/user/1000/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

func (f *DefaultClientFactory) NewPodmanSystemClient() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

// ContainerEngine implements container engine detection and management with auto-detection.
type ContainerEngine struct {
	Client client.APIClient
	EngineName   string
}

// NewContainerEngine creates a new container engine with auto-detection.
// It tries to connect to a container engine and returns the first available one.
func NewContainerEngine() (*ContainerEngine, error) {
	return newContainerEngineWithFactory(&DefaultClientFactory{})
}

// NewContainerEngineWithFactory creates a container engine using the provided factory.
// This is used internally and for testing.
func NewContainerEngineWithFactory(factory ClientFactory) (*ContainerEngine, error) {
	return newContainerEngineWithFactory(factory)
}

// newContainerEngineWithFactory is the internal implementation
func newContainerEngineWithFactory(factory ClientFactory) (*ContainerEngine, error) {
	ctx := context.Background()
	
	// Try Docker first (most common)
	dockerClient, err := factory.NewDockerClient()
	if err == nil {
		engine := &ContainerEngine{
			Client: dockerClient,
			EngineName:   "Docker",
		}
		if ready, _ := engine.CheckReady(ctx); ready {
			return engine, nil
		}
	}

	// Try Podman with Docker-compatible socket
	podmanClient, err := factory.NewPodmanUserClient()
	if err == nil {
		engine := &ContainerEngine{
			Client: podmanClient,
			EngineName:   "Podman",
		}
		if ready, _ := engine.CheckReady(ctx); ready {
			return engine, nil
		}
	}

	// Try system-wide Podman socket
	podmanSystemClient, err := factory.NewPodmanSystemClient()
	if err == nil {
		engine := &ContainerEngine{
			Client: podmanSystemClient,
			EngineName:   "Podman",
		}
		if ready, _ := engine.CheckReady(ctx); ready {
			return engine, nil
		}
	}

	return nil, ErrNoContainerEngine
}

// CheckReady checks if the container engine is available using the API client.
func (u *ContainerEngine) CheckReady(ctx context.Context) (bool, error) {
	_, err := u.Client.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("%s ping failed: %w", u.EngineName, err)
	}

	return true, nil
}

// GetName returns the name of the detected container engine.
func (u *ContainerEngine) GetName() string {
	return u.EngineName
}


