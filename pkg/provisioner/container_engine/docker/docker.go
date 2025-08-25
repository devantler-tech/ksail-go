package dockerprovisioner

import (
	"context"
	"github.com/docker/docker/client"
)

// DockerProvisioner implements ContainerEngineProvisioner for Docker.
type DockerProvisioner struct {
	client client.APIClient
}

// CheckReady checks if the Docker service/socket is available using the Docker API client.
func (d *DockerProvisioner) CheckReady() (bool, error) {
	ctx := context.Background()

	_, err := d.client.Ping(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

// NewDockerProvisioner creates a new DockerProvisioner.
func NewDockerProvisioner() *DockerProvisioner {
	cli, _ := client.NewClientWithOpts(client.FromEnv)

	return &DockerProvisioner{client: cli}
}
