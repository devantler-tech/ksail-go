# pkg/provisioner/containerengine

This package provides container engine management and abstraction for KSail.

## Purpose

Provides abstraction and management for different container engines (Docker, Podman) used by KSail. This package handles the detection, initialization, and interaction with container engines required for running Kubernetes clusters locally.

## Features

- **Multi-Engine Support**: Supports Docker and Podman container engines
- **Auto-Detection**: Automatically detects available container engines
- **Fallback Logic**: Tries multiple engines in order of preference
- **Engine Abstraction**: Provides a common interface for different engines
- **Health Checking**: Verifies container engine readiness and availability

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/provisioner/containerengine"

// Create container engine with auto-detection
engine, err := containerengine.NewContainerEngine()
if err != nil {
    log.Fatal("Failed to initialize container engine:", err)
}

// Check if engine is ready
if err := engine.CheckReady(); err != nil {
    log.Fatal("Container engine not ready:", err)
}

// Get engine name
engineName := engine.Name()
fmt.Printf("Using container engine: %s\n", engineName)
```

This package is essential for KSail's local cluster provisioning, as it provides the foundation for container-based Kubernetes distributions like Kind and K3d.

---

[⬅️ Go Back](../provisioner/README.md)
