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

// ClientFactory defines the interface for creating container engine clients.
type ClientFactory interface {
	GetDockerClient() (client.APIClient, error)
	GetPodmanUserClient() (client.APIClient, error)
	GetPodmanSystemClient() (client.APIClient, error)
}

// DefaultClientFactory provides the default implementation for client creation.
type DefaultClientFactory struct{}

// GetDockerClient creates a Docker client using environment configuration.
func (f *DefaultClientFactory) GetDockerClient() (client.APIClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// GetPodmanUserClient creates a Podman client using the user-specific socket.
func (f *DefaultClientFactory) GetPodmanUserClient() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/user/1000/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

// GetPodmanSystemClient creates a Podman client using the system-wide socket.
func (f *DefaultClientFactory) GetPodmanSystemClient() (client.APIClient, error) {
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

// NewContainerEngine creates a new container engine with optional dependency injection.
// If apiClient is nil, auto-detection is performed to find an available container engine.
// If apiClient is provided, it creates an engine with the injected client and engineName.
func NewContainerEngine(apiClient client.APIClient, engineName string) (*ContainerEngine, error) {
	// If client is provided, use dependency injection
	if apiClient != nil {
		return &ContainerEngine{
			Client:     apiClient,
			EngineName: engineName,
		}, nil
	}

	// Auto-detection mode when client is nil
	return GetAutoDetectedClient()
}

// GetAutoDetectedClient attempts to auto-detect and create a container engine client.
// It tries Docker first, then Podman with different socket configurations.
func GetAutoDetectedClient() (*ContainerEngine, error) {
	factory := &DefaultClientFactory{}
	return GetAutoDetectedClientWithFactory(factory)
}

// GetAutoDetectedClientWithFactory attempts to auto-detect and create a container engine client
// using the provided factory. This allows for dependency injection and testing.
func GetAutoDetectedClientWithFactory(factory ClientFactory) (*ContainerEngine, error) {
	ctx := context.Background()
	
	// Try Docker first (most common)
	dockerClient, err := factory.GetDockerClient()
	if err == nil {
		engine := &ContainerEngine{
			Client:     dockerClient,
			EngineName: "Docker",
		}
		if ready, _ := engine.CheckReady(ctx); ready {
			return engine, nil
		}
	}

	// Try Podman with Docker-compatible socket
	podmanClient, err := factory.GetPodmanUserClient()
	if err == nil {
		engine := &ContainerEngine{
			Client:     podmanClient,
			EngineName: "Podman",
		}
		if ready, _ := engine.CheckReady(ctx); ready {
			return engine, nil
		}
	}

	// Try system-wide Podman socket
	podmanSystemClient, err := factory.GetPodmanSystemClient()
	if err == nil {
		engine := &ContainerEngine{
			Client:     podmanSystemClient,
			EngineName: "Podman",
		}
		if ready, _ := engine.CheckReady(ctx); ready {
			return engine, nil
		}
	}

	return nil, ErrNoContainerEngine
}

// GetDockerClient creates a Docker client using environment configuration.
func GetDockerClient() (client.APIClient, error) {
	factory := &DefaultClientFactory{}
	return factory.GetDockerClient()
}

// GetPodmanUserClient creates a Podman client using the user-specific socket.
func GetPodmanUserClient() (client.APIClient, error) {
	factory := &DefaultClientFactory{}
	return factory.GetPodmanUserClient()
}

// GetPodmanSystemClient creates a Podman client using the system-wide socket.
func GetPodmanSystemClient() (client.APIClient, error) {
	factory := &DefaultClientFactory{}
	return factory.GetPodmanSystemClient()
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


