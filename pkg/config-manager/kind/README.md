# pkg/config-manager/kind

This package provides configuration management for Kind cluster configurations.

## Purpose

Implements file-based configuration loading for Kind clusters using the `sigs.k8s.io/kind/pkg/apis/config/v1alpha4.Cluster` type. Provides simple YAML file loading with automatic Kind defaults application.

## Features

- **File-based Loading**: Load Kind cluster configurations from YAML files
- **Default Configuration**: Returns sensible defaults when configuration file doesn't exist
- **Path Traversal**: For relative paths, searches up the directory tree to find configuration files
- **Caching**: Loads configuration once and caches for subsequent calls
- **Kind Defaults**: Automatically applies Kind's built-in defaults via `v1alpha4.SetDefaultsCluster`
- **TypeMeta Completion**: Ensures proper APIVersion and Kind fields are set

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"

// Create a config manager pointing to a Kind config file
manager := kind.NewConfigManager("/path/to/kind-config.yaml")

// Load the configuration (file or defaults)
config, err := manager.LoadConfig()
if err != nil {
    log.Fatal("Failed to load Kind config:", err)
}

// Use the loaded Kind cluster configuration
fmt.Printf("Cluster: %s (API: %s)\n", config.Kind, config.APIVersion)
fmt.Printf("Nodes: %d\n", len(config.Nodes))
```

## Configuration File Format

The Kind config manager expects standard Kind cluster configuration files:

```yaml
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: my-cluster
nodes:
- role: control-plane
  kubeadmConfigPatches:
  - |
    kind: InitConfiguration
    nodeRegistration:
      kubeletExtraArgs:
        node-labels: "ingress-ready=true"
  extraPortMappings:
  - containerPort: 80
    hostPort: 80
    protocol: TCP
- role: worker
- role: worker
```

## Behavior

- **Absolute Paths**: Uses the path directly to look for the configuration file
- **Relative Paths**: Starts from current directory and traverses up the tree looking for the file
- **File Exists**: Loads and parses the YAML file, applies Kind defaults
- **File Missing**: Returns a default Kind cluster configuration with one control-plane node
- **Invalid YAML**: Returns an error with details about the parsing failure
- **Caching**: Subsequent calls to `LoadConfig()` return the cached configuration

## Interface Compliance

This package implements the `configmanager.ConfigManager[v1alpha4.Cluster]` interface:

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

Run tests with:

```bash
go test ./pkg/config-manager/kind/...
```

---

[⬅️ Go Back](../README.md)
