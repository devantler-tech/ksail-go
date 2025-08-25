// Package podmanprovisioner provides a Podman implementation of the ContainerEngineProvisioner interface.
package podmanprovisioner

import (
	"context"
	"fmt"
	"os"

	"github.com/docker/docker/client"
)

// PodmanProvisioner implements ContainerEngineProvisioner for Podman.
type PodmanProvisioner struct {
	client client.APIClient
}

// NewPodmanProvisioner creates a new PodmanProvisioner with a provided client.
func NewPodmanProvisioner(client client.APIClient) *PodmanProvisioner {
	podmanSock := fmt.Sprintf("unix:///run/user/%d/podman/podman.sock", os.Getuid())

	err := os.Setenv("DOCKER_HOST", podmanSock)
	if err != nil {
		panic(fmt.Sprintf("failed to set DOCKER_HOST: %v", err))
	}

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
