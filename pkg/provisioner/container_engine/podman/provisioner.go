// Package podmanprovisioner provides a Podman implementation of the ContainerEngineProvisioner interface.
package podmanprovisioner

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// PodmanProvisioner implements ContainerEngineProvisioner for Podman.
type PodmanProvisioner struct {
	client client.APIClient
}

// NewPodmanProvisioner creates a new PodmanProvisioner with a provided client.
func NewPodmanProvisioner(client client.APIClient) (*PodmanProvisioner) {
	return &PodmanProvisioner{client: client}
}

// CheckReady checks if the Podman service/socket is available.
func (p *PodmanProvisioner) CheckReady() (bool, error) {
	ctx := context.Background()

	_, err := p.client.Ping(ctx)
	if err != nil {
		return false, fmt.Errorf("podman ping failed: %w", err)
	}

	return true, nil
}
