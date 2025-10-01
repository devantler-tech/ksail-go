# pkg/scaffolder

This package provides scaffolding utilities for KSail projects.

## Purpose

Contains utilities for generating KSail project files and configurations. The scaffold creator provides the necessary configuration files for a complete KSail project setup.

## Features

- **Multi-Distribution Support**: Scaffolds configurations for different Kubernetes distributions (Kind, K3d, EKS)
- **Complete Project Setup**: Generates all required files including ksail.yaml, distribution configs, and kustomization files
- **Force Overwrite**: Option to overwrite existing files
- **Directory Structure**: Creates proper directory structure for source files

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/scaffolder"

// Create a cluster configuration
cluster := v1alpha1.Cluster{
    TypeMeta: metav1.TypeMeta{
        APIVersion: v1alpha1.APIVersion,
        Kind:       v1alpha1.Kind,
    },
    Metadata: metav1.ObjectMeta{
        Name: "my-cluster",
    },
    Spec: v1alpha1.Spec{
        Distribution:       v1alpha1.DistributionKind,
        SourceDirectory:    "k8s",
        DistributionConfig: "kind.yaml",
    },
}

// Create scaffold instance and generate files
scaffold := scaffolder.NewScaffolder(cluster, os.Stdout)
err := scaffold.Scaffold("/path/to/output/", false)
if err != nil {
    log.Fatal(err)
}
```

## Generated Files

The scaffold creator generates the following files based on the cluster configuration:

- **ksail.yaml**: Main KSail cluster configuration
- **Distribution config**: Kind, K3d, or EKS-specific configuration file
- **kustomization.yaml**: Kustomization file in the source directory

### Distribution-Specific Files

- **Kind**: Generates `kind.yaml` with a basic cluster configuration
- **K3d**: Generates `k3d.yaml` with a simple k3d cluster configuration
- **EKS**: Generates `eks.yaml` with an EKS cluster configuration including node groups

---

[⬅️ Go Back](../README.md)
