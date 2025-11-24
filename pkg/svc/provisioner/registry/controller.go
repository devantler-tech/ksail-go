package registry

import (
	"context"
	"fmt"
	"io"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/client"
)

// Controller orchestrates registry lifecycle operations shared between mirror and local registries.
type Controller struct {
	backend Backend
}

// NewController creates a controller backed by the provided registry backend.
func NewController(backend Backend) (*Controller, error) {
	if backend == nil {
		return nil, fmt.Errorf("registry backend is required")
	}

	return &Controller{backend: backend}, nil
}

// EnsureBatch creates all requested registries as an atomic batch. Any failure rolls back prior creations.
func (c *Controller) EnsureBatch(
	ctx context.Context,
	registries []Info,
	clusterName string,
	networkName string,
	writer io.Writer,
) error {
	if len(registries) == 0 {
		return nil
	}

	batch, err := newMirrorBatch(ctx, c.backend, clusterName, networkName, writer, len(registries))
	if err != nil {
		return err
	}

	for _, reg := range registries {
		if _, err := batch.ensure(ctx, reg); err != nil {
			batch.rollback(ctx)
			return err
		}
	}

	return nil
}

// EnsureOne provisions a single registry and reports whether a new container was created.
func (c *Controller) EnsureOne(
	ctx context.Context,
	spec Info,
	clusterName string,
	writer io.Writer,
) (bool, error) {
	tracker, err := newMirrorBatch(ctx, c.backend, clusterName, "", writer, 1)
	if err != nil {
		return false, err
	}

	created, ensureErr := tracker.ensure(ctx, spec)
	if ensureErr != nil {
		tracker.rollback(ctx)
	}

	return created, ensureErr
}

// Cleanup removes the provided registries via the backend.
func (c *Controller) Cleanup(
	ctx context.Context,
	registries []Info,
	clusterName string,
	deleteVolumes bool,
	networkName string,
	writer io.Writer,
) error {
	return CleanupRegistries(ctx, c.backend, registries, clusterName, deleteVolumes, networkName, writer)
}

// CleanupOne removes a single registry spec.
func (c *Controller) CleanupOne(
	ctx context.Context,
	registry Info,
	clusterName string,
	deleteVolume bool,
	networkName string,
	writer io.Writer,
) error {
	return c.backend.DeleteRegistry(ctx, registry.Name, clusterName, deleteVolume, networkName)
}

// NewDockerController constructs a controller backed by the Docker RegistryManager.
func NewDockerController(apiClient client.APIClient) (*Controller, error) {
	mgr, err := dockerclient.NewRegistryManager(apiClient)
	if err != nil {
		return nil, err
	}

	return NewController(mgr)
}
