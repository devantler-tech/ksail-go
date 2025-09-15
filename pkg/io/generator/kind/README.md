# pkg/io/generator/kind

This package provides Kind-specific resource generators for KSail.

## Purpose

Generates Kubernetes resources and configuration files specifically tailored for Kind (Kubernetes in Docker) clusters. Kind is a tool for running local Kubernetes clusters using Docker container "nodes".

## Features

- **Kind-Specific Resources**: Generates resources optimized for Kind environments
- **Docker Integration**: Handles Docker-specific configurations for Kind clusters
- **Local Development**: Generates configurations suitable for local Kubernetes development
- **Multi-Node Support**: Supports configurations for single and multi-node Kind clusters

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"

// Generate Kind-specific resources
generator := kind.NewGenerator(/* configuration */)
resources, err := generator.Generate(/* parameters */)
if err != nil {
    log.Fatal("Failed to generate Kind resources:", err)
}
```

This generator is used when KSail needs to create resources specifically for Kind clusters, ensuring compatibility with Kind's Docker-based node architecture and local development workflows.

---

[⬅️ Go Back](../README.md)
