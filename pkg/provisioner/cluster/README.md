# cluster

This package provides core cluster provisioning implementations for KSail.

## Purpose

Implements the `ClusterProvisioner` interface for managing Kubernetes clusters across different distributions. This package contains the core logic for cluster lifecycle operations including creation, deletion, starting, stopping, and listing clusters.

## Interface

```go
type ClusterProvisioner interface {
    // Create creates a Kubernetes cluster
    Create(ctx context.Context, name string) error
    
    // Delete deletes a Kubernetes cluster by name
    Delete(ctx context.Context, name string) error
    
    // Start starts a Kubernetes cluster by name
    Start(ctx context.Context, name string) error
    
    // Stop stops a Kubernetes cluster by name
    Stop(ctx context.Context, name string) error
    
    // List lists all Kubernetes clusters
    List(ctx context.Context) ([]string, error)
    
    // Exists checks if a Kubernetes cluster exists by name
    Exists(ctx context.Context, name string) (bool, error)
}
```

## Features

- **Context Support**: All operations support context for cancellation and timeouts
- **Multi-Distribution**: Support for different Kubernetes distributions
- **Lifecycle Management**: Complete cluster lifecycle operations
- **Name-based Operations**: Cluster identification by name

## Subpackages

- `eks/` - Amazon EKS cluster provisioning
- `k3d/` - K3d cluster provisioning
- `kind/` - Kind cluster provisioning

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"

// Use specific cluster provisioner
var provisioner ClusterProvisioner = // ... get specific implementation

ctx := context.Background()

// Create a cluster
if err := provisioner.Create(ctx, "my-cluster"); err != nil {
    log.Fatal("Failed to create cluster:", err)
}

// List clusters
clusters, err := provisioner.List(ctx)
if err != nil {
    log.Fatal("Failed to list clusters:", err)
}
```

---

[⬅️ Go Back](../README.md)