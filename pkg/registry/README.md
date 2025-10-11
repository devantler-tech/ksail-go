# pkg/registry

This package provides functionality for managing Docker registry containers used as mirror registries for Kind and K3d clusters.

## Purpose

Implements Docker registry container lifecycle management and configuration parsing for mirror registries. When Kind or K3d clusters are configured with registry mirrors, this package ensures those mirror registry containers are created before the cluster.

## Features

- **Registry Container Management**: Create and manage Docker registry containers
- **Configuration Parsing**: Extract registry configurations from Kind containerd patches and K3d registry configs
- **Automatic Creation**: Automatically create mirror registries before cluster creation
- **Idempotent Operations**: Skip creation if registry containers already exist
- **Default Image Support**: Uses `registry:3` Docker image by default

## Usage

### Creating a Registry Manager

```go
import (
    "github.com/devantler-tech/ksail-go/pkg/registry"
    "github.com/docker/docker/client"
)

// Create a Docker client
dockerClient, err := client.NewClientWithOpts(client.FromEnv)
if err != nil {
    log.Fatal(err)
}

// Create registry manager
manager := registry.NewManager(dockerClient)
```

### Creating a Registry Container

```go
ctx := context.Background()

cfg := registry.RegistryConfig{
    Name:     "k3d-registry",
    HostPort: "5000",
    Image:    registry.DefaultRegistryImage, // or specify custom image
}

err := manager.CreateRegistry(ctx, cfg)
if err != nil {
    log.Fatal("Failed to create registry:", err)
}
```

### Extracting Registries from K3d Config

```go
import k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"

k3dConfig := &k3dv1alpha5.SimpleConfig{
    Registries: k3dv1alpha5.SimpleConfigRegistries{
        Use: []string{"k3d-registry:5000"},
    },
}

registries, err := registry.ExtractRegistriesFromK3d(k3dConfig)
if err != nil {
    log.Fatal(err)
}

// Create each registry
for _, reg := range registries {
    err := manager.CreateRegistry(ctx, reg)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Extracting Registries from Kind Config

```go
import "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"

kindConfig := &v1alpha4.Cluster{
    ContainerdConfigPatches: []string{`
[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
  endpoint = ["http://kind-registry:5000"]
`},
}

registries, err := registry.ExtractRegistriesFromKind(kindConfig)
if err != nil {
    log.Fatal(err)
}

// Create each registry
for _, reg := range registries {
    err := manager.CreateRegistry(ctx, reg)
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration Formats

### K3d Registry Configuration

K3d uses the `Registries.Use` field to specify registries:

```yaml
apiVersion: k3d.io/v1alpha5
kind: Simple
metadata:
  name: my-cluster
registries:
  use:
    - k3d-registry:5000
```

The parser extracts the name and port from the registry reference format: `<name>:<port>`

### Kind Registry Configuration

Kind uses containerd configuration patches to configure registry mirrors:

```yaml
apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
      endpoint = ["http://kind-registry:5000"]
```

The parser extracts:
- Mirror host from the `mirrors."<host>:<port>"` section
- Registry name and port from the `endpoint` URL

## Integration with Provisioners

The registry package is integrated into the Kind provisioner. When a Kind cluster is created, the provisioner:

1. Parses the Kind configuration for containerd patches
2. Extracts any registry mirror configurations
3. Creates the required registry containers
4. Proceeds with cluster creation

For K3d, registry creation is handled natively by K3d itself, so manual creation is typically not needed.

## Testing

The package includes comprehensive unit tests:

```bash
go test ./pkg/registry/...
```

Tests cover:
- Registry manager operations (create, exists checking)
- Configuration parsing for both Kind and K3d
- Error handling and edge cases
- Default image usage

## Error Handling

The package provides detailed error messages for:
- Registry container creation failures
- Image pull failures
- Configuration parsing errors
- Invalid registry reference formats

## Design Decisions

- **Interface-Based**: Uses `ContainerAPIClient` interface for testability
- **Idempotent**: Safe to call CreateRegistry multiple times
- **Default Image**: Uses `registry:3` as the default registry image
- **Automatic Port Binding**: Binds to host port specified in configuration
- **Restart Policy**: Sets `always` restart policy for registry containers

---

[⬅️ Go Back](../../README.md)
