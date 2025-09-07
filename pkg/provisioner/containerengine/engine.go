// Package containerengine provides unified container engine detection and management.
package containerengine

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/client"
)

// Error definitions for container engine operations.
var (
	// ErrNoContainerEngine is returned when no container engine (Docker or Podman) is available.
	ErrNoContainerEngine = errors.New("no container engine (Docker or Podman) available")
	// ErrAPIClientNil is returned when apiClient is nil.
	ErrAPIClientNil = errors.New("apiClient cannot be nil")
	// ErrClientNotReady is returned when a client is not ready.
	ErrClientNotReady = errors.New("client not ready")
	// ErrEngineDetection is returned when engine type cannot be detected.
	ErrEngineDetection = errors.New("unable to detect engine type from client")
)

// ContainerEngine implements container engine detection and management.
type ContainerEngine struct {
	Client client.APIClient
}

// NewContainerEngine creates a new container engine with dependency injection.
// The apiClient must be provided - this function detects the engine type from the client.
func NewContainerEngine(apiClient client.APIClient) (*ContainerEngine, error) {
	if apiClient == nil {
		return nil, ErrAPIClientNil
	}

	return &ContainerEngine{
		Client: apiClient,
	}, nil
}

// ClientCreator is a function type for creating container engine clients.
type ClientCreator func() (client.APIClient, error)

// detectEngineType detects the container engine type from the client.
func (u *ContainerEngine) detectEngineType(ctx context.Context) (string, error) {
	version, err := u.Client.ServerVersion(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get server version: %w", err)
	}

	// Check platform name to determine engine type
	platformName := version.Platform.Name
	if platformName != "" {
		// Docker typically returns "Docker Engine - Community" or similar
		if contains(platformName, "Docker") {
			return "Docker", nil
		}
		// Podman typically returns something with "Podman" in the name
		if contains(platformName, "Podman") {
			return "Podman", nil
		}
	}

	// Fallback: check version string
	versionStr := version.Version
	if versionStr != "" {
		if contains(versionStr, "podman") {
			return "Podman", nil
		}
		// If it doesn't contain "podman", assume Docker as it's more common
		return "Docker", nil
	}

	return "", ErrEngineDetection
}

// contains is a helper function for case-insensitive string matching.
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
// GetAutoDetectedClient attempts to automatically detect and create a container engine client.
// It tries Docker first, then Podman with different socket configurations.
// For testing, you can override specific creators using a map with keys:
// "docker", "podman-user", "podman-system".
func GetAutoDetectedClient(overrides ...map[string]ClientCreator) (*ContainerEngine, error) {
	// Default client creators
	creators := map[string]ClientCreator{
		"docker":        GetDockerClient,
		"podman-user":   GetPodmanUserClient,
		"podman-system": GetPodmanSystemClient,
	}

	// Apply overrides for testing
	if len(overrides) > 0 && overrides[0] != nil {
		for key, creator := range overrides[0] {
			creators[key] = creator
		}
	}

	ctx := context.Background()

	// Try Docker first (most common)
	if engine, err := tryCreateEngine(ctx, creators["docker"]); err == nil {
		return engine, nil
	}

	// Try Podman with Docker-compatible socket
	if engine, err := tryCreateEngine(ctx, creators["podman-user"]); err == nil {
		return engine, nil
	}

	// Try system-wide Podman socket
	if engine, err := tryCreateEngine(ctx, creators["podman-system"]); err == nil {
		return engine, nil
	}

	return nil, ErrNoContainerEngine
}

// tryCreateEngine attempts to create and validate a container engine.
func tryCreateEngine(ctx context.Context, creator ClientCreator) (*ContainerEngine, error) {
	apiClient, err := creator()
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	engine, err := NewContainerEngine(apiClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create container engine: %w", err)
	}

	if ready, _ := engine.CheckReady(ctx); ready {
		return engine, nil
	}

	return nil, fmt.Errorf("%w: client not ready", ErrClientNotReady)
}


// ClientCreatorFunc defines a function type for creating container engine clients.
//
// Deprecated: Use ClientCreator instead.
type ClientCreatorFunc func() (client.APIClient, error)

// DefaultDockerClientCreator is the default implementation for creating Docker clients.
var DefaultDockerClientCreator ClientCreator = func() (client.APIClient, error) {
	return client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
}

// DefaultPodmanUserClientCreator is the default implementation for creating Podman user clients.
var DefaultPodmanUserClientCreator ClientCreator = func() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/user/1000/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

// DefaultPodmanSystemClientCreator is the default implementation for creating Podman system clients.
var DefaultPodmanSystemClientCreator ClientCreator = func() (client.APIClient, error) {
	return client.NewClientWithOpts(
		client.WithHost("unix:///run/podman/podman.sock"),
		client.WithAPIVersionNegotiation(),
	)
}

// GetDockerClient creates a Docker client using environment configuration.
func GetDockerClient() (client.APIClient, error) {
	dockerClient, err := DefaultDockerClientCreator()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	return dockerClient, nil
}

// GetPodmanUserClient creates a Podman client using the user-specific socket.
func GetPodmanUserClient() (client.APIClient, error) {
	podmanClient, err := DefaultPodmanUserClientCreator()
	if err != nil {
		return nil, fmt.Errorf("failed to create Podman user client: %w", err)
	}

	return podmanClient, nil
}

// GetPodmanSystemClient creates a Podman client using the system-wide socket.
func GetPodmanSystemClient() (client.APIClient, error) {
	podmanClient, err := DefaultPodmanSystemClientCreator()
	if err != nil {
		return nil, fmt.Errorf("failed to create Podman system client: %w", err)
	}

	return podmanClient, nil
}

// CheckReady checks if the container engine is available using the API client.
func (u *ContainerEngine) CheckReady(ctx context.Context) (bool, error) {
	_, err := u.Client.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("container engine ping failed: %w", err)
	}

	return true, nil
}

// GetName returns the name of the detected container engine.
func (u *ContainerEngine) GetName() string {
	ctx := context.Background()

	engineType, err := u.detectEngineType(ctx)
	if err != nil {
		// Fallback to "Unknown" if detection fails
		return "Unknown"
	}

	return engineType
}


