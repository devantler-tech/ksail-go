# pkg/io/generator/k3d

This package provides K3d-specific resource generators for KSail.

## Purpose

Generates Kubernetes resources and configuration files specifically tailored for K3d clusters. K3d is a lightweight wrapper to run k3s (Rancher Lab's minimal Kubernetes distribution) in Docker.

## Features

- **K3d-Specific Resources**: Generates resources optimized for K3d environments
- **Docker Integration**: Handles Docker-specific configurations for K3d clusters
- **Lightweight Setup**: Generates configurations suitable for local development with K3d
- **k3s Compatibility**: Ensures generated resources work with the k3s distribution

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"

// Generate K3d-specific resources
generator := k3d.NewGenerator(/* configuration */)
resources, err := generator.Generate(/* parameters */)
if err != nil {
    log.Fatal("Failed to generate K3d resources:", err)
}
```

This generator is used when KSail needs to create resources specifically for K3d clusters, ensuring compatibility with K3d's Docker-based architecture and k3s features.

---

[⬅️ Go Back](../README.md)
