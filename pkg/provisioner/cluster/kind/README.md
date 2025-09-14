# kind

This package provides Kind cluster provisioning for KSail.

## Purpose

Implements the `ClusterProvisioner` interface specifically for Kind (Kubernetes in Docker) clusters. Kind is a tool for running local Kubernetes clusters using Docker container "nodes", primarily designed for testing Kubernetes itself.

## Features

- **Kind Integration**: Native integration with Kind tooling
- **Docker Backend**: Uses Docker containers as Kubernetes nodes
- **Multi-Node Support**: Supports single and multi-node cluster configurations
- **Local Development**: Optimized for local Kubernetes development and testing
- **Standard Kubernetes**: Runs a full Kubernetes distribution (not a minimal one)

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"

// Create Kind provisioner
kindProvisioner := kind.NewProvisioner(/* Kind configuration */)

ctx := context.Background()

// Create Kind cluster
if err := kindProvisioner.Create(ctx, "my-kind-cluster"); err != nil {
    log.Fatal("Failed to create Kind cluster:", err)
}

// List Kind clusters
clusters, err := kindProvisioner.List(ctx)
if err != nil {
    log.Fatal("Failed to list clusters:", err)
}

// Delete cluster
if err := kindProvisioner.Delete(ctx, "my-kind-cluster"); err != nil {
    log.Fatal("Failed to delete Kind cluster:", err)
}
```

This provisioner is excellent for local development and testing scenarios where a full Kubernetes distribution is needed in a containerized environment.