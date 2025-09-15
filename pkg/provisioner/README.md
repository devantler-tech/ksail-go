# pkg/provisioner

This package provides cluster provisioning functionality for KSail.

## Purpose

Contains the core provisioning logic and interfaces for managing Kubernetes clusters across different providers and distributions. This package serves as the foundation for KSail's cluster lifecycle management capabilities.

## Features

- **Multi-Provider Support**: Supports different Kubernetes distributions and cloud providers
- **Lifecycle Management**: Create, delete, start, stop, and list clusters
- **Provider Abstraction**: Common interface across different cluster providers
- **Container Engine Integration**: Works with different container engines (Docker, Podman)

## Subpackages

- **[cluster/](./cluster/README.md)** - Core cluster provisioning implementations
- **[containerengine/](./containerengine/README.md)** - Container engine management and abstraction

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/provisioner"

// Use cluster provisioner implementations
// See individual subpackage documentation for detailed usage
```

The provisioner package provides a unified approach to managing Kubernetes clusters regardless of the underlying infrastructure or distribution, enabling KSail to work consistently across different environments.

---

[⬅️ Go Back](../README.md)