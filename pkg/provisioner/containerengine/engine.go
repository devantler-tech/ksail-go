// Package containerengine provides unified container engine detection and management.
package containerengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine/factory"
	"github.com/docker/docker/client"
)

// ErrNoContainerEngine is returned when no container engine (Docker or Podman) is available.
var ErrNoContainerEngine = errors.New("no container engine (Docker or Podman) available")

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
// An optional factory can be provided for dependency injection and testing.
func GetAutoDetectedClient(factories ...factory.ClientFactory) (*ContainerEngine, error) {
	var clientFactory factory.ClientFactory
	if len(factories) > 0 {
		clientFactory = factories[0]
	} else {
		clientFactory = &factory.DefaultClientFactory{}
	}

	ctx := context.Background()
	
	// Try Docker first (most common)
	dockerClient, err := clientFactory.GetDockerClient()
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
	podmanClient, err := clientFactory.GetPodmanUserClient()
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
	podmanSystemClient, err := clientFactory.GetPodmanSystemClient()
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
	clientFactory := &factory.DefaultClientFactory{}
	return clientFactory.GetDockerClient()
}

// GetPodmanUserClient creates a Podman client using the user-specific socket.
func GetPodmanUserClient() (client.APIClient, error) {
	clientFactory := &factory.DefaultClientFactory{}
	return clientFactory.GetPodmanUserClient()
}

// GetPodmanSystemClient creates a Podman client using the system-wide socket.
func GetPodmanSystemClient() (client.APIClient, error) {
	clientFactory := &factory.DefaultClientFactory{}
	return clientFactory.GetPodmanSystemClient()
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


