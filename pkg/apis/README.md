# apis

This directory contains Kubernetes API definitions for KSail.

## Purpose

Houses the Kubernetes API definitions, custom resource types, and related schemas used by KSail. This directory follows Kubernetes API conventions for versioning and organization.

## Features

- **Custom Resource Definitions**: KSail-specific Kubernetes resource types
- **API Versioning**: Proper versioning following Kubernetes conventions
- **Schema Definitions**: Type-safe API definitions for KSail resources
- **Kubernetes Integration**: Native integration with Kubernetes API machinery

## Structure

- **[cluster/](./cluster/README.md)** - Cluster-related API definitions organized by version

## Usage

These API definitions are used for creating and managing KSail's custom Kubernetes resources:

```go
import "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"

// Use KSail's custom resource types
// See specific version directories for detailed usage
```

The APIs follow Kubernetes conventions for versioning (v1alpha1, v1beta1, v1, etc.) and provide strongly-typed definitions for KSail's Kubernetes resources.

---

[⬅️ Go Back](../README.md)