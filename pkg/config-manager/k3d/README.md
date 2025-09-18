# pkg/config-manager/k3d

This package provides configuration management for K3d cluster configurations.

## Purpose

Implements file-based configuration loading for K3d clusters using the `github.com/k3d-io/k3d/v5/pkg/config/v1alpha5.SimpleConfig` type. Provides simple YAML file loading with automatic defaults application.

## Features

- **File-based Loading**: Load K3d cluster configurations from YAML files
- **Default Configuration**: Returns sensible defaults when configuration file doesn't exist
- **Path Traversal**: For relative paths, searches up the directory tree to find configuration files
- **Caching**: Loads configuration once and caches for subsequent calls
- **TypeMeta Completion**: Ensures proper APIVersion and Kind fields are set
- **Safe File Reading**: Uses secure file reading utilities to prevent path traversal attacks

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"

// Create a config manager pointing to a K3d config file
manager := k3d.NewConfigManager("/path/to/k3d-config.yaml")

// Load the configuration (file or defaults)
config, err := manager.LoadConfig()
if err != nil {
    log.Fatal("Failed to load K3d config:", err)
}

// Use the loaded K3d cluster configuration
fmt.Printf("Cluster: %s (API: %s)\n", config.Kind, config.APIVersion)
fmt.Printf("Servers: %d, Agents: %d\n", config.Servers, config.Agents)
```

## Configuration File Format

The K3d config manager expects standard K3d cluster configuration files:

```yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: my-cluster
servers: 1
agents: 2
image: rancher/k3s:latest
network: k3d-network
clusterToken: my-token
volumes:
  - volume: /tmp/data:/data
    nodeFilters:
      - server:0
ports:
  - port: 8080:80
    nodeFilters:
      - loadbalancer
```

## Behavior

- **Absolute Paths**: Uses the path directly to look for the configuration file
- **Relative Paths**: Starts from current directory and traverses up the tree looking for the file
- **File Exists**: Loads and parses the YAML file, ensures TypeMeta fields are set
- **File Missing**: Returns a default K3d cluster configuration
- **Invalid YAML**: Returns an error with details about the parsing failure
- **Caching**: Subsequent calls to `LoadConfig()` return the cached configuration

## Interface Compliance

This package implements the `configmanager.ConfigManager[v1alpha5.SimpleConfig]` interface:

```go
type ConfigManager[T any] interface {
    LoadConfig() (*T, error)
}
```

## Testing

The package includes comprehensive tests covering:

- File loading with valid configurations
- Default behavior when file doesn't exist
- Error handling for invalid YAML
- Configuration caching
- TypeMeta completion
- Path traversal functionality

Run tests with:

```bash
go test ./pkg/config-manager/k3d/...
```

---

[⬅️ Go Back](../README.md)