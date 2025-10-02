# pkg/provisioner/cluster/kind

This package provides Kind cluster provisioning for KSail.

## Purpose

Implements the `ClusterProvisioner` interface specifically for Kind (Kubernetes in Docker) clusters. Kind is a tool for running local Kubernetes clusters using Docker container "nodes", primarily designed for testing Kubernetes itself.

## Features

- **Kind Integration**: Native integration with Kind tooling
- **Docker Backend**: Uses Docker containers as Kubernetes nodes
- **Multi-Node Support**: Supports single and multi-node cluster configurations
- **Local Development**: Optimized for local Kubernetes development and testing
- **Standard Kubernetes**: Runs a full Kubernetes distribution (not a minimal one)
- **Production Adapters**: Built-in production-ready adapters for Kind provider and Docker client

## Usage

### With Production Adapters (Recommended)

```go
import (
    kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
    "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Create production-ready adapters
providerAdapter := kindprovisioner.NewDefaultKindProviderAdapter()
dockerClient, err := kindprovisioner.NewDefaultDockerClient()
if err != nil {
    log.Fatal("Failed to create Docker client:", err)
}

// Create a kind configuration
kindConfig := &v1alpha4.Cluster{}
kindConfig.Name = "my-kind-cluster"

// Create Kind provisioner with adapters
kindProvisioner := kindprovisioner.NewKindClusterProvisioner(
    kindConfig,
    "~/.kube/config",
    providerAdapter,
    dockerClient,
)

ctx := context.Background()

// Create Kind cluster
if err := kindProvisioner.Create(ctx, ""); err != nil {
    log.Fatal("Failed to create Kind cluster:", err)
}

// List Kind clusters
clusters, err := kindProvisioner.List(ctx)
if err != nil {
    log.Fatal("Failed to list clusters:", err)
}

// Delete cluster
if err := kindProvisioner.Delete(ctx, ""); err != nil {
    log.Fatal("Failed to delete Kind cluster:", err)
}
```

### Custom Adapters for Testing

For testing, you can provide custom implementations of `KindProvider` and `ContainerAPIClient` interfaces to mock Kind and Docker behavior.

This provisioner is excellent for local development and testing scenarios where a full Kubernetes distribution is needed in a containerized environment.

---

[⬅️ Go Back](../README.md)
