# pkg/io/generator

This package provides resource generation utilities for KSail.

## Purpose

Contains utilities for generating Kubernetes resources and configuration files for different cluster distributions and components. The generators create declarative configurations that can be applied to Kubernetes clusters.

## Features

- **Multi-Distribution Support**: Generators for different Kubernetes distributions
- **Resource Generation**: Creates Kubernetes manifests and configuration files
- **Declarative Output**: Generates configurations in standard Kubernetes formats

## Subpackages

- **[eks/](./eks/README.md)** - Amazon EKS-specific resource generators
- **[k3d/](./k3d/README.md)** - K3d-specific resource generators  
- **[kind/](./kind/README.md)** - Kind-specific resource generators
- **[kustomization/](./kustomization/README.md)** - Kustomization file generators
- **[yaml/](./yaml/README.md)** - Generic YAML resource generators

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/generator"

// Use specific generators for different distributions
// See individual subpackage documentation for detailed usage
```

Each subpackage provides specialized generators for their respective target platforms, enabling KSail to generate appropriate configurations for different Kubernetes environments.

---

[⬅️ Go Back](../README.md)