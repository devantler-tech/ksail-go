# pkg/installer/helm

This package provides a Helm installer implementation for KSail.

## Purpose

Implements the `Installer` interface specifically for installing and managing Helm charts. This installer enables users to deploy workloads to Kubernetes clusters using Helm charts through KSail's unified CLI interface.

## Features

- **Helm Chart Installation**: Installs Helm charts from various sources (repositories, OCI registries)
- **Version Management**: Supports specific chart versions
- **Namespace Configuration**: Configures target namespace for chart deployment
- **Values Customization**: Supports custom values via values files
- **Timeout Configuration**: Configurable timeout for installation operations
- **Atomic Operations**: Uses atomic flag for safe rollbacks on failure

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/installer/helm"

// Create helm client
client, err := helmclient.NewClientFromKubeConf(&helmclient.KubeConfClientOptions{
    Options: &helmclient.Options{
        Namespace: "default",
    },
    KubeConfig: kubeconfigBytes,
})
if err != nil {
    log.Fatal("Failed to create helm client:", err)
}

// Create helm installer
helmInstaller := helm.NewHelmInstaller(
    client,
    "my-release",              // Release name
    "stable/nginx-ingress",    // Chart name
    "default",                 // Namespace
    "1.2.3",                   // Version (optional)
    valuesYaml,                // Values YAML content (optional)
    5*time.Minute,             // Timeout
)

// Install chart
ctx := context.Background()
if err := helmInstaller.Install(ctx); err != nil {
    log.Fatal("Failed to install chart:", err)
}

// Uninstall chart
if err := helmInstaller.Uninstall(ctx); err != nil {
    log.Fatal("Failed to uninstall chart:", err)
}
```

## CLI Integration

The helm installer is integrated into the `ksail workload install` command:

```bash
# Install a chart from a repository
ksail workload install my-release stable/nginx-ingress

# Install with specific version
ksail workload install my-release stable/nginx-ingress --version 1.2.3

# Install with custom values
ksail workload install my-release stable/nginx-ingress --values values.yaml

# Install an OCI chart
ksail workload install my-release oci://registry/repo/chart
```

This installer is essential for deploying workloads to Kubernetes clusters managed by KSail.

---

[⬅️ Go Back](../README.md)
