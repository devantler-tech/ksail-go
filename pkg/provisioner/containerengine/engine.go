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

// ContainerEngine implements container engine detection and management.
type ContainerEngine struct {
	Client     client.APIClient
	EngineName string
}

// NewContainerEngine creates a new container engine with dependency injection.
// The apiClient and engineName must be provided - this function no longer performs auto-detection.
// For auto-detection, use GetAutoDetectedClient() separately.
func NewContainerEngine(apiClient client.APIClient, engineName string) (*ContainerEngine, error) {
	if apiClient == nil {
		return nil, errors.New("apiClient cannot be nil - use GetAutoDetectedClient() for auto-detection")
	}
	if engineName == "" {
		return nil, errors.New("engineName cannot be empty")
	}

	return &ContainerEngine{
		Client:     apiClient,
		EngineName: engineName,
	}, nil
}

// GetAutoDetectedClient attempts to auto-detect and create a container engine client.
// It tries Docker first, then Podman with different socket configurations.
// Optional client creators can be provided for dependency injection and testing
// using the ClientCreators struct with named fields.
// If no creators are provided, the default static functions are used.
func GetAutoDetectedClient(creators ...*ClientCreators) (*ContainerEngine, error) {
	dockerCreator := GetDockerClient
	podmanUserCreator := GetPodmanUserClient  
	podmanSystemCreator := GetPodmanSystemClient

	// Override with provided creators for testing (use first non-nil creators struct)
	if len(creators) > 0 && creators[0] != nil {
		clientCreators := creators[0]
		if clientCreators.Docker != nil {
			dockerCreator = clientCreators.Docker
		}
		if clientCreators.PodmanUser != nil {
			podmanUserCreator = clientCreators.PodmanUser
		}
		if clientCreators.PodmanSystem != nil {
			podmanSystemCreator = clientCreators.PodmanSystem
		}
	}

	ctx := context.Background()
	
	// Try Docker first (most common)
	dockerClient, err := dockerCreator()
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
	podmanClient, err := podmanUserCreator()
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
	podmanSystemClient, err := podmanSystemCreator()
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
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// GetPodmanUserClient creates a Podman client using the user-specific socket.
func GetPodmanUserClient() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/user/1000/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

// GetPodmanSystemClient creates a Podman client using the system-wide socket.
func GetPodmanSystemClient() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
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


