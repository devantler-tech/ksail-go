# pkg/apis/cluster

This directory contains cluster-related API definitions for KSail.

## Purpose

Contains the Kubernetes API definitions and custom resource types specifically related to cluster management. These APIs define the schema and types used for KSail's cluster lifecycle operations.

## Features

- **Cluster Resource Types**: Custom resource definitions for cluster management
- **Versioned APIs**: Multiple API versions following Kubernetes conventions
- **Type Safety**: Strongly-typed Go definitions for cluster resources
- **Kubernetes Native**: Full integration with Kubernetes API machinery

## Versions

- **[pkg/apis/cluster/v1alpha1/](./v1alpha1/README.md)** - Alpha version of cluster APIs (development/testing)

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"

// Use cluster API types for KSail cluster management
// See specific version directories for detailed type definitions
```

The cluster APIs provide the foundation for KSail's declarative cluster management capabilities, allowing users to define cluster configurations as Kubernetes resources.

---

[⬅️ Go Back](../README.md)