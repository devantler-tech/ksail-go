// Package dockerprovisioner provides a Docker implementation of the ContainerEngineProvisioner interface.
package dockerprovisioner

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// DockerProvisioner implements ContainerEngineProvisioner for Docker.
type DockerProvisioner struct {
	client client.APIClient
}

// NewDockerProvisioner creates a new DockerProvisioner with a provided client.
func NewDockerProvisioner(client client.APIClient) *DockerProvisioner {
	return &DockerProvisioner{client: client}
}

// CheckReady checks if the Docker service/socket is available using the Docker API client.
func (d *DockerProvisioner) CheckReady() (bool, error) {
	ctx := context.Background()

	_, err := d.client.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("docker ping failed: %w", err)
	}

	return true, nil
}
