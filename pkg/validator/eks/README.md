# pkg/validator/eks

This package provides validation functionality for EKS cluster configurations.

## Purpose

Contains validation logic specifically for Amazon EKS cluster configurations. This validator ensures EKS configurations are valid according to eksctl APIs and Amazon EKS requirements before cluster provisioning.

## Features

- **EKS Configuration Validation**: Validates EKS cluster configurations using upstream eksctl APIs
- **Metadata Validation**: Integrates with metadata validator for consistent validation
- **AWS Best Practices**: Enforces AWS and EKS best practices and requirements
- **Error Reporting**: Detailed validation error messages and reporting
- **Pre-provisioning Validation**: Catches configuration errors before attempting cluster creation

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator/eks"

// Create EKS validator
validator := &eks.Validator{}

// Validate EKS cluster configuration
result := validator.Validate(eksConfig)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Validation error: %v\n", err)
    }
}
```

## Key Components

- **validator.go**: Core EKS validation logic and implementation
- **eksctl integration**: Uses upstream eksctl APIs for validation
- **metadata validation**: Leverages shared metadata validation utilities

## Integration

This validator is used by:

- `pkg/config-manager/eks`: EKS configuration loading and validation
- EKS cluster provisioning workflows
- CLI commands that work with EKS configurations

## Validation Rules

The EKS validator enforces:

- Valid EKS cluster naming conventions
- Proper AWS region and availability zone configurations
- Valid instance types and node group configurations
- Kubernetes version compatibility
- IAM role and policy requirements
- Network and security group configurations

This ensures EKS configurations are valid before attempting expensive cloud provisioning operations.

---

[⬅️ Go Back](../../README.md)
