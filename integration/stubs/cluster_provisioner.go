// Package stubs provides stub implementations of core interfaces for integration testing.
//
// These stubs allow integration tests to verify command behavior and workflows
// without requiring actual cluster provisioning or file I/O operations.
// All stub implementations output their actions to help with debugging test failures.
package stubs

import (
	"context"
	"fmt"
)

// ClusterProvisioner is a stub implementation of the ClusterProvisioner interface for integration testing.
// It provides no-op implementations that only print stub messages.
type ClusterProvisioner struct {
	ClusterName string
}

// NewClusterProvisioner creates a new stub cluster provisioner.
func NewClusterProvisioner(clusterName string) *ClusterProvisioner {
	return &ClusterProvisioner{
		ClusterName: clusterName,
	}
}

// Create simulates creating a Kubernetes cluster.
func (p *ClusterProvisioner) Create(_ context.Context, name string) error {
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Create cluster (name=%s)\n", name)

	return nil
}

// Delete simulates deleting a Kubernetes cluster.
func (p *ClusterProvisioner) Delete(_ context.Context, name string) error {
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Delete cluster (name=%s)\n", name)

	return nil
}

// Start simulates starting a Kubernetes cluster.
func (p *ClusterProvisioner) Start(_ context.Context, name string) error {
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Start cluster (name=%s)\n", name)

	return nil
}

// Stop simulates stopping a Kubernetes cluster.
func (p *ClusterProvisioner) Stop(_ context.Context, name string) error {
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Stop cluster (name=%s)\n", name)

	return nil
}

// List simulates listing all Kubernetes clusters.
func (p *ClusterProvisioner) List(_ context.Context) ([]string, error) {
	//nolint:forbidigo // Using fmt.Println for test stub output
	fmt.Println("STUB: List clusters")

	return []string{}, nil
}

// Exists simulates checking if a Kubernetes cluster exists.
func (p *ClusterProvisioner) Exists(_ context.Context, name string) (bool, error) {
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Check cluster exists (name=%s)\n", name)

	return true, nil
}
