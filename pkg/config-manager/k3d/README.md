# pkg/config-manager/k3d

This package provides configuration management for K3d v1alpha5.SimpleConfig configurations.

## Purpose

Implements simple file-based configuration loading for K3d clusters without Viper dependency. This allows loading K3d configuration files directly from disk and unmarshalling them into the appropriate K3d types.

## Features

- **File-based Loading**: Load K3d configurations directly from YAML files
- **K3d Compatibility**: Works with K3d v1alpha5.SimpleConfig format
- **No Viper Dependency**: Simple file-based approach without configuration management framework
- **Caching**: Loaded configurations are cached to avoid repeated file I/O
- **Type Safety**: Properly typed interface compliance with ConfigManager generic interface

## Usage

```go
import k3dconfig "github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"

// Create config manager for a specific file
manager := k3dconfig.NewConfigManager("/path/to/k3d-config.yaml")

// Load the configuration
config, err := manager.LoadConfig()
if err != nil {
    log.Fatal("Failed to load K3d config:", err)
}

// Use the loaded K3d SimpleConfig
fmt.Printf("Cluster name: %s\n", config.Name)
fmt.Printf("Servers: %d, Agents: %d\n", config.Servers, config.Agents)
```

## Supported Configuration Format

The manager supports the standard K3d v1alpha5 SimpleConfig format:

```yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: my-cluster
servers: 1
agents: 2
image: rancher/k3s:latest
network: custom-network
clusterToken: my-token
```

## Interface Compliance

This config manager implements the `configmanager.ConfigManager[v1alpha5.SimpleConfig]` interface, ensuring compatibility with the broader KSail configuration management system.

---

[⬅️ Go Back](../README.md)