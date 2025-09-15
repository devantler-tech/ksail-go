# pkg/provisioner/cluster/eks

This package provides Amazon EKS cluster provisioning for KSail.

## Purpose

Implements the `ClusterProvisioner` interface specifically for Amazon Elastic Kubernetes Service (EKS) clusters. This provisioner handles the creation, management, and lifecycle operations for EKS clusters on AWS.

## Features

- **EKS Integration**: Native integration with Amazon EKS service
- **AWS Authentication**: Handles AWS authentication and authorization
- **Node Group Management**: Manages EKS node groups and worker nodes
- **VPC Integration**: Works with AWS VPC networking
- **IAM Integration**: Handles IAM roles and policies for EKS

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/eks"

// Create EKS provisioner
eksProvisioner := eks.NewProvisioner(/* AWS configuration */)

ctx := context.Background()

// Create EKS cluster
if err := eksProvisioner.Create(ctx, "my-eks-cluster"); err != nil {
    log.Fatal("Failed to create EKS cluster:", err)
}

// Check if cluster exists
exists, err := eksProvisioner.Exists(ctx, "my-eks-cluster")
if err != nil {
    log.Fatal("Failed to check cluster existence:", err)
}
```

This provisioner is used when KSail needs to manage production-ready Kubernetes clusters on Amazon Web Services using the EKS service.

---

[⬅️ Go Back](../README.md)