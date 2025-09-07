// Package factory provides container engine client factory interfaces and implementations.
package factory

import (
	"github.com/docker/docker/client"
)

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