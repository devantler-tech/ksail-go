package clusterprovisioner

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

// DockerClient describes the subset of methods from Docker's API used here.
type DockerClient interface {
	ContainerStart(ctx context.Context, name string, options container.StartOptions) error
	ContainerStop(ctx context.Context, name string, options container.StopOptions) error
}
