# flux

This package provides a Flux installer implementation for KSail.

## Purpose

Implements the `Installer` interface specifically for installing and managing Flux, a GitOps toolkit for Kubernetes. This installer handles the deployment and removal of Flux components in Kubernetes clusters.

## Features

- **Flux Installation**: Installs Flux components to Kubernetes clusters
- **Flux Removal**: Cleanly uninstalls Flux and its resources
- **GitOps Integration**: Enables GitOps workflows for cluster management

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/installer/flux"

// Create Flux installer
fluxInstaller := flux.NewInstaller(/* configuration */)

// Install Flux
ctx := context.Background()
if err := fluxInstaller.Install(ctx); err != nil {
    log.Fatal("Failed to install Flux:", err)
}

// Uninstall Flux
if err := fluxInstaller.Uninstall(ctx); err != nil {
    log.Fatal("Failed to uninstall Flux:", err)
}
```

This installer is typically used when setting up GitOps workflows for Kubernetes cluster management with KSail.

---

[⬅️ Go Back](../README.md)