# Quickstart: KSail Project Scaffolder

## Overview

The KSail Project Scaffolder generates minimal configuration files for new KSail projects, enabling quick setup of local Kubernetes development environments.

## Basic Usage

### Step 1: Create Cluster Configuration

```go
import (
    "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
    "github.com/devantler-tech/ksail-go/pkg/scaffolder"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Define your cluster configuration
cluster := v1alpha1.Cluster{
    TypeMeta: metav1.TypeMeta{
        APIVersion: v1alpha1.APIVersion,
        Kind:       v1alpha1.Kind,
    },
    Metadata: metav1.ObjectMeta{
        Name: "my-dev-cluster",
    },
    Spec: v1alpha1.Spec{
        Distribution:       v1alpha1.DistributionKind,
        SourceDirectory:    "k8s",
        DistributionConfig: "kind.yaml",
    },
}
```

### Step 2: Initialize Scaffolder

```go
// Create scaffolder instance
scaffolder := scaffolder.NewScaffolder(cluster)
```

### Step 3: Generate Project Files

```go
// Generate all project files
err := scaffolder.Scaffold("/path/to/project/", false)
if err != nil {
    log.Fatal(err)
}
```

## Generated File Structure

After scaffolding, your project will contain:

```text
my-project/
├── ksail.yaml          # Main KSail configuration
├── kind.yaml           # Kind cluster configuration
└── k8s/
    └── kustomization.yaml  # Kubernetes resource management
```

## Distribution Support

### Kind (Local Docker)

```go
cluster.Spec.Distribution = v1alpha1.DistributionKind
cluster.Spec.DistributionConfig = "kind.yaml"
```

Generates minimal Kind cluster configuration suitable for local development.

### K3d (Lightweight Kubernetes)

```go
cluster.Spec.Distribution = v1alpha1.DistributionK3d
cluster.Spec.DistributionConfig = "k3d.yaml"
```

Generates K3d configuration for lightweight Kubernetes clusters.

### AWS EKS (Cloud)

```go
cluster.Spec.Distribution = v1alpha1.DistributionEKS
cluster.Spec.DistributionConfig = "eks.yaml"
```

Generates EKS configuration with sensible defaults for cloud deployment.

## Common Patterns

### Force Overwrite Existing Files

```go
err := scaffolder.Scaffold("/path/to/project/", true) // force = true
```

### Different Source Directories

```go
cluster.Spec.SourceDirectory = "manifests"
// Will create manifests/kustomization.yaml instead of k8s/
```

### Custom Distribution Config Names

```go
cluster.Spec.DistributionConfig = "my-kind-config.yaml"
// Will generate my-kind-config.yaml instead of kind.yaml
```

## Error Handling

```go
err := scaffolder.Scaffold(outputPath, force)
if err != nil {
    switch {
    case errors.Is(err, scaffolder.ErrTindNotImplemented):
        log.Println("Tind distribution not yet supported")
    case errors.Is(err, scaffolder.ErrUnknownDistribution):
        log.Println("Unknown distribution specified")
    default:
        log.Printf("Scaffolding failed: %v", err)
    }
}
```

## Integration with KSail CLI

The scaffolder is typically used through the KSail CLI:

```bash
ksail init --distribution kind --name my-cluster
```

This uses the scaffolder internally to generate all necessary project files.

## Next Steps

After scaffolding:

1. **Review Generated Files**: Examine ksail.yaml and distribution config
2. **Customize Configuration**: Modify files according to your needs
3. **Start Development**: Use `ksail up` to create your cluster
4. **Add Resources**: Place Kubernetes manifests in the source directory
