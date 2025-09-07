// Package containerengine provides unified container engine detection and management.
package containerengine

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/client"
)

// Error definitions for container engine operations.
var (
	// ErrNoContainerEngine is returned when no container engine (Docker or Podman) is available.
	ErrNoContainerEngine = errors.New("no container engine (Docker or Podman) available")
	// ErrAPIClientNil is returned when apiClient is nil.
	ErrAPIClientNil = errors.New("apiClient cannot be nil - use GetAutoDetectedClient() for auto-detection")
	// ErrEngineNameEmpty is returned when engineName is empty.
	ErrEngineNameEmpty = errors.New("engineName cannot be empty")
	// ErrClientNotReady is returned when a client is not ready.
	ErrClientNotReady = errors.New("client not ready")
)

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
		return nil, ErrAPIClientNil
	}

	if engineName == "" {
		return nil, ErrEngineNameEmpty
	}

	return &ContainerEngine{
		Client:     apiClient,
		EngineName: engineName,
	}, nil
}

// ClientCreator is a function type for creating container engine clients.
type ClientCreator func() (client.APIClient, error)

// GetAutoDetectedClient attempts to auto-detect and create a container engine client.
// It tries Docker first, then Podman with different socket configurations.
// For testing, you can override specific creators using a map with keys:
// "docker", "podman-user", "podman-system".
func GetAutoDetectedClient(overrides ...map[string]ClientCreator) (*ContainerEngine, error) {
	creators := getClientCreators(overrides...)
	
	ctx := context.Background()
	
	// Try Docker first (most common)
	engine, err := tryCreateEngine(ctx, creators.docker, "Docker")
	if err == nil {
		return engine, nil
	}

	// Try Podman with Docker-compatible socket
	engine, err = tryCreateEngine(ctx, creators.podmanUser, "Podman")
	if err == nil {
		return engine, nil
	}

	// Try system-wide Podman socket
	engine, err = tryCreateEngine(ctx, creators.podmanSystem, "Podman")
	if err == nil {
		return engine, nil
	}

	return nil, ErrNoContainerEngine
}

// clientCreators holds the client creation functions.
type clientCreators struct {
	docker        ClientCreator
	podmanUser    ClientCreator
	podmanSystem  ClientCreator
}

// getClientCreators returns the client creators with optional overrides.
func getClientCreators(overrides ...map[string]ClientCreator) clientCreators {
	creators := clientCreators{
		docker:        GetDockerClient,
		podmanUser:    GetPodmanUserClient,
		podmanSystem:  GetPodmanSystemClient,
	}

	// Override with provided creators for testing
	if len(overrides) > 0 && overrides[0] != nil {
		overrideMap := overrides[0]

		if creator, exists := overrideMap["docker"]; exists {
			creators.docker = creator
		}

		if creator, exists := overrideMap["podman-user"]; exists {
			creators.podmanUser = creator
		}

		if creator, exists := overrideMap["podman-system"]; exists {
			creators.podmanSystem = creator
		}
	}

	return creators
}

// tryCreateEngine attempts to create and validate a container engine.
func tryCreateEngine(ctx context.Context, creator ClientCreator, engineName string) (*ContainerEngine, error) {
	apiClient, err := creator()
	if err != nil {
		return nil, fmt.Errorf("failed to create %s client: %w", engineName, err)
	}

	engine := &ContainerEngine{
		Client:     apiClient,
		EngineName: engineName,
	}

	if ready, _ := engine.CheckReady(ctx); ready {
		return engine, nil
	}

	return nil, fmt.Errorf("%w: %s", ErrClientNotReady, engineName)
}


// GetDockerClient creates a Docker client using environment configuration.
func GetDockerClient() (client.APIClient, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return dockerClient, nil
}

// GetPodmanUserClient creates a Podman client using the user-specific socket.
func GetPodmanUserClient() (client.APIClient, error) {
	podmanClient, err := client.NewClientWithOpts(
		client.WithHost("unix:///run/user/1000/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Podman user client: %w", err)
	}

	return podmanClient, nil
}

// GetPodmanSystemClient creates a Podman client using the system-wide socket.
func GetPodmanSystemClient() (client.APIClient, error) {
	podmanClient, err := client.NewClientWithOpts(
		client.WithHost("unix:///run/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Podman system client: %w", err)
	}

	return podmanClient, nil
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


