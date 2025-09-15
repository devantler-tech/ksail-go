# internal/utils/k8s

This package provides Kubernetes utilities for KSail's internal use.

## Purpose

Contains internal utility functions and helpers for working with Kubernetes clusters, resources, and APIs. These utilities are used internally by KSail components and are not intended for external consumption.

## Features

- **Kubernetes Integration**: Utilities for interacting with Kubernetes APIs
- **Resource Management**: Helpers for managing Kubernetes resources
- **Cluster Operations**: Internal utilities for cluster-related operations
- **API Abstractions**: Simplified interfaces for common Kubernetes operations

## Usage

This package is for internal use within KSail. It provides common Kubernetes-related functionality that is shared across different KSail components.

```go
import "github.com/devantler-tech/ksail-go/internal/utils/k8s"

// Internal usage within KSail components
// Not intended for external consumption
```

**Note**: This is an internal package and its API may change without notice. External applications should not depend on this package directly.

---

[⬅️ Go Back](../README.md)