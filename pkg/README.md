# pkg

This directory contains the core business logic packages for KSail.

## Purpose

The `pkg/` directory houses all the main functionality packages that implement KSail's core features. These packages provide the underlying business logic for cluster provisioning, configuration management, file I/O operations, and component installation.

## Features

- **Public API**: All packages in `pkg/` are part of KSail's public API and can be imported by external applications
- **Modular Design**: Each package focuses on a specific domain area with clear responsibilities
- **Production Ready**: Comprehensive testing and documentation for reliable usage
- **Cross-Platform**: Designed to work across different operating systems and environments

## Packages

- **[pkg/apis/](./apis/cluster/v1alpha1/README.md)** - Kubernetes API definitions and custom resource types
- **[pkg/config-manager/](./config-manager/README.md)** - Configuration management utilities for handling KSail and application configs
- **[pkg/installer/](./installer/README.md)** - Component installation utilities (kubectl, Flux, etc.)
- **[pkg/io/](./io/README.md)** - Safe file I/O operations with security features and resource generation
- **[pkg/provisioner/](./provisioner/README.md)** - Cluster provisioning and lifecycle management across different Kubernetes distributions

## Architecture

The packages in `pkg/` follow clean architecture principles:

- **Domain Separation**: Each package addresses a specific domain concern
- **Interface-Based Design**: Heavy use of interfaces for testability and extensibility  
- **Dependency Injection**: Components accept their dependencies as interfaces
- **Context Support**: All long-running operations support context for cancellation and timeouts

## Usage

```go
import (
    "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
    "github.com/devantler-tech/ksail-go/pkg/config-manager"
    "github.com/devantler-tech/ksail-go/pkg/io"
)

// Use packages for building KSail functionality
// See individual package README files for detailed usage examples
```

Each package provides comprehensive documentation, examples, and interfaces for building robust Kubernetes cluster management applications.

---

[⬅️ Go Back](../README.md)
