# pkg/installer/kubectl

This package provides a kubectl installer implementation for KSail.

## Purpose

Implements the `Installer` interface specifically for installing and managing kubectl, the Kubernetes command-line tool. This installer ensures that kubectl is available and properly configured for use with KSail.

## Features

- **kubectl Installation**: Downloads and installs kubectl binary
- **Version Management**: Handles specific kubectl versions
- **Configuration**: Sets up kubectl configuration for cluster access

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/installer/kubectl"

// Create kubectl installer
kubectlInstaller := kubectl.NewInstaller(/* configuration */)

// Install kubectl
ctx := context.Background()
if err := kubectlInstaller.Install(ctx); err != nil {
    log.Fatal("Failed to install kubectl:", err)
}

// Uninstall kubectl
if err := kubectlInstaller.Uninstall(ctx); err != nil {
    log.Fatal("Failed to uninstall kubectl:", err)
}
```

This installer is essential for ensuring kubectl availability when working with Kubernetes clusters through KSail.

---

[⬅️ Go Back](../README.md)
