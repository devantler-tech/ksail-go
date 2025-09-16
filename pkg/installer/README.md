# pkg/installer

This package provides functionality for installing and uninstalling components in KSail.

## Purpose

Defines the core `Installer` interface and provides implementations for installing various components required by KSail, such as CLI tools and Kubernetes resources.

## Interface

```go
type Installer interface {
    // Install installs the component
    Install(ctx context.Context) error
    
    // Uninstall uninstalls the component
    Uninstall(ctx context.Context) error
}
```

## Features

- **Context Support**: All operations support context for cancellation and timeouts
- **Uniform Interface**: Consistent interface across all installer implementations
- **Error Handling**: Proper error reporting for installation/uninstallation operations

## Subpackages

- **[pkg/installer/flux/](./flux/README.md)** - Flux installer implementation
- **[pkg/installer/kubectl/](./kubectl/README.md)** - Kubectl installer implementation

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/installer"

// Use any installer implementation
var installer Installer = // ... get specific installer

// Install component
ctx := context.Background()
if err := installer.Install(ctx); err != nil {
    log.Fatal(err)
}

// Uninstall component
if err := installer.Uninstall(ctx); err != nil {
    log.Fatal(err)
}
```

---

[⬅️ Go Back](../README.md)
