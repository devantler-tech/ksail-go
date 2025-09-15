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

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"

// Create K3d provisioner
k3dProvisioner := k3d.NewProvisioner(/* K3d configuration */)

ctx := context.Background()

// Create K3d cluster
if err := k3dProvisioner.Create(ctx, "my-k3d-cluster"); err != nil {
    log.Fatal("Failed to create K3d cluster:", err)
}

// Start cluster
if err := k3dProvisioner.Start(ctx, "my-k3d-cluster"); err != nil {
    log.Fatal("Failed to start K3d cluster:", err)
}

// Stop cluster
if err := k3dProvisioner.Stop(ctx, "my-k3d-cluster"); err != nil {
    log.Fatal("Failed to stop K3d cluster:", err)
}
```

This provisioner is ideal for local development workflows where quick, lightweight Kubernetes clusters are needed for testing and development.

---

[⬅️ Go Back](../README.md)