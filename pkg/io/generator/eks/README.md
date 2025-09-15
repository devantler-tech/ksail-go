# eks

This package provides Amazon EKS-specific resource generators for KSail.

## Purpose

Generates Kubernetes resources and configuration files specifically tailored for Amazon Elastic Kubernetes Service (EKS) clusters. This includes EKS-specific manifests, configurations, and deployment files.

## Features

- **EKS-Specific Resources**: Generates resources optimized for EKS environments
- **AWS Integration**: Handles AWS-specific configurations and integrations
- **IAM Considerations**: Generates resources that work with AWS IAM and security models
- **EKS Add-ons**: Supports generation for EKS-specific add-ons and features

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"

// Generate EKS-specific resources
generator := eks.NewGenerator(/* configuration */)
resources, err := generator.Generate(/* parameters */)
if err != nil {
    log.Fatal("Failed to generate EKS resources:", err)
}
```

This generator is used when KSail needs to create resources specifically for Amazon EKS clusters, ensuring compatibility with AWS services and EKS-specific features.

---

[⬅️ Go Back](../../../../README.md)