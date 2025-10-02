# pkg/provisioner/cluster/k3d

This package provides K3d cluster provisioning for KSail.

## Purpose

Implements the `ClusterProvisioner` interface specifically for K3d clusters. K3d is a lightweight wrapper to run k3s (Rancher Lab's minimal Kubernetes distribution) in Docker, making it ideal for local development and testing.

## Features

- **K3d Integration**: Native integration with K3d tooling
- **Docker Backend**: Uses Docker containers as cluster nodes
- **Fast Setup**: Quick cluster creation and teardown for development
- **Resource Efficient**: Lightweight k3s distribution optimized for development
- **Port Mapping**: Supports port mapping for local access to services
- **Production Adapters**: Built-in production-ready adapters for k3d client and config operations

## Usage

### With Production Adapters (Recommended)

```go
import (
    k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
    v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// Create production-ready adapters
clientAdapter := k3dprovisioner.NewDefaultK3dClientAdapter()
configAdapter := k3dprovisioner.NewDefaultK3dConfigAdapter()

// Create a simple configuration
simpleCfg := &v1alpha5.SimpleConfig{}
simpleCfg.Name = "my-k3d-cluster"

// Create K3d provisioner with adapters
k3dProvisioner := k3dprovisioner.NewK3dClusterProvisioner(
    simpleCfg,
    clientAdapter,
    configAdapter,
)

ctx := context.Background()

// Create K3d cluster
if err := k3dProvisioner.Create(ctx, ""); err != nil {
    log.Fatal("Failed to create K3d cluster:", err)
}

// Start cluster
if err := k3dProvisioner.Start(ctx, ""); err != nil {
    log.Fatal("Failed to start K3d cluster:", err)
}

// Stop cluster
if err := k3dProvisioner.Stop(ctx, ""); err != nil {
    log.Fatal("Failed to stop K3d cluster:", err)
}
```

### Custom Adapters for Testing

For testing, you can provide custom implementations of `K3dClientProvider` and `K3dConfigProvider` interfaces to mock k3d behavior.

This provisioner is ideal for local development workflows where quick, lightweight Kubernetes clusters are needed for testing and development.

---

[⬅️ Go Back](../README.md)
