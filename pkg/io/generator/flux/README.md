# Flux Resource Generators

This package provides generators for creating Flux CD resources using the official flux2 manifestgen library.

## Overview

The package contains two main generators:

1. **InstallGenerator** - Generates Flux installation manifests
2. **SyncGenerator** - Generates GitRepository and Kustomization resources for continuous deployment

## Usage Examples

### FluxInstallGenerator

Generate Flux installation manifests for deploying Flux to a cluster:

```go
package main

import (
    "fmt"
    "time"

    fluxgen "github.com/devantler-tech/ksail-go/pkg/io/generator/flux"
    "github.com/fluxcd/flux2/v2/pkg/manifestgen/install"
)

func main() {
    // Create generator
    gen := fluxgen.NewInstallGenerator()

    // Configure installation options
    installOpts := install.MakeDefaultOptions()
    installOpts.Version = "v2.7.3"
    installOpts.Namespace = "flux-system"
    installOpts.Components = []string{"source-controller", "kustomize-controller", "helm-controller"}
    installOpts.Timeout = 5 * time.Minute

    // Generate manifest (returns YAML string)
    opts := fluxgen.InstallOptions{
        Options: installOpts,
    }
    manifest, err := gen.Generate(nil, opts)
    if err != nil {
        panic(err)
    }
    fmt.Println(manifest)

    // Or write directly to file
    opts.Output = "./flux-install.yaml"
    opts.Force = true
    result, err := gen.Generate(nil, opts)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Written to: %s\n", result)
}
```

### FluxSyncGenerator

Generate GitRepository and Kustomization resources for continuous deployment:

```go
package main

import (
    "fmt"
    "time"

    fluxgen "github.com/devantler-tech/ksail-go/pkg/io/generator/flux"
    "github.com/fluxcd/flux2/v2/pkg/manifestgen/sync"
)

func main() {
    // Create generator
    gen := fluxgen.NewSyncGenerator()

    // Configure sync options
    syncOpts := sync.Options{
        Name:       "my-app",
        Namespace:  "flux-system",
        URL:        "https://github.com/myorg/myrepo",
        Branch:     "main",
        Interval:   5 * time.Minute,
        TargetPath: "./clusters/production",
        Secret:     "git-credentials",
    }

    // Generate manifest (returns YAML string)
    opts := fluxgen.SyncOptions{
        Options: syncOpts,
    }
    manifest, err := gen.Generate(nil, opts)
    if err != nil {
        panic(err)
    }
    fmt.Println(manifest)

    // Or write directly to file
    opts.Output = "./flux-sync.yaml"
    opts.Force = true
    result, err := gen.Generate(nil, opts)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Written to: %s\n", result)
}
```

### Using Different Git References

The SyncGenerator supports multiple Git reference types:

```go
// Branch reference
syncOpts.Branch = "main"

// Tag reference
syncOpts.Tag = "v1.0.0"

// Commit reference
syncOpts.Commit = "abc123def456"

// Semantic version reference
syncOpts.SemVer = ">=1.0.0 <2.0.0"
```

## Features

- **Flux Installation**: Generate complete Flux installation manifests from GitHub releases or local paths
- **GitOps Sync**: Create GitRepository and Kustomization resources for continuous deployment
- **Flexible Output**: Return manifests as strings or write directly to files
- **Force Overwrite**: Optional force flag for overwriting existing files
- **Full Options**: All flux2 manifestgen options are supported through embedded structs

## Integration with KSail

These generators follow KSail's generator pattern:
- Implement the `Generator[T, Options]` interface pattern
- Use KSail's `io.TryWriteFile()` for safe file operations
- Include comprehensive error handling and validation

## Dependencies

- [fluxcd/flux2](https://github.com/fluxcd/flux2) - Official Flux CD CLI and manifestgen library
