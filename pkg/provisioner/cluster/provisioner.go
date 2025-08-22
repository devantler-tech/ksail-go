// Package clusterprovisioner provides implementations of the Provisioner interface
// for provisioning clusters in different providers.
package clusterprovisioner

import "context"

// ClusterProvisioner defines methods for managing Kubernetes clusters.
type ClusterProvisioner interface {
	// Create creates a Kubernetes cluster. If name is non-empty, target that name; otherwise use config defaults.
	Create(name string, ctx context.Context) error

	// Delete deletes a Kubernetes cluster by name or config default when name is empty.
	Delete(name string, ctx context.Context) error

	// Start starts a Kubernetes cluster by name or config default when name is empty.
	Start(name string, ctx context.Context) error

	// Stop stops a Kubernetes cluster by name or config default when name is empty.
	Stop(name string, ctx context.Context) error

	// List lists all Kubernetes clusters.
	List(ctx context.Context) ([]string, error)

	// Exists checks if a Kubernetes cluster exists by name or config default when name is empty.
	Exists(name string, ctx context.Context) (bool, error)
}
