package stubs

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

// DockerClientStub is a stub implementation of Docker container API client.
type DockerClientStub struct {
	ContainerStartError error
	ContainerStopError  error

	StartCalls []string
	StopCalls  []string
}

// NewDockerClientStub creates a new DockerClientStub.
func NewDockerClientStub() *DockerClientStub {
	return &DockerClientStub{}
}

// ContainerStart simulates container start.
func (d *DockerClientStub) ContainerStart(
	ctx context.Context,
	containerID string,
	options container.StartOptions,
) error {
	d.StartCalls = append(d.StartCalls, containerID)
	return d.ContainerStartError
}

// ContainerStop simulates container stop.
func (d *DockerClientStub) ContainerStop(
	ctx context.Context,
	containerID string,
	options container.StopOptions,
) error {
	d.StopCalls = append(d.StopCalls, containerID)
	return d.ContainerStopError
}
