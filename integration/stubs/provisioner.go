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
	clusterName := name
	if clusterName == "" {
		clusterName = p.ClusterName
	}
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Creating cluster '%s'\n", clusterName)

	return nil
}

// Delete simulates deleting a Kubernetes cluster.
func (p *ClusterProvisioner) Delete(_ context.Context, name string) error {
	clusterName := name
	if clusterName == "" {
		clusterName = p.ClusterName
	}
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Deleting cluster '%s'\n", clusterName)

	return nil
}

// Start simulates starting a Kubernetes cluster.
func (p *ClusterProvisioner) Start(_ context.Context, name string) error {
	clusterName := name
	if clusterName == "" {
		clusterName = p.ClusterName
	}
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Starting cluster '%s'\n", clusterName)

	return nil
}

// Stop simulates stopping a Kubernetes cluster.
func (p *ClusterProvisioner) Stop(_ context.Context, name string) error {
	clusterName := name
	if clusterName == "" {
		clusterName = p.ClusterName
	}
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Stopping cluster '%s'\n", clusterName)

	return nil
}

// List simulates listing all Kubernetes clusters.
func (p *ClusterProvisioner) List(_ context.Context) ([]string, error) {
	//nolint:forbidigo // Using fmt.Println for test stub output
	fmt.Println("STUB: Listing clusters")

	return []string{p.ClusterName}, nil
}

// Exists simulates checking if a Kubernetes cluster exists.
func (p *ClusterProvisioner) Exists(_ context.Context, name string) (bool, error) {
	clusterName := name
	if clusterName == "" {
		clusterName = p.ClusterName
	}
	//nolint:forbidigo // Using fmt.Printf for test stub output
	fmt.Printf("STUB: Checking if cluster '%s' exists\n", clusterName)

	return true, nil
}
