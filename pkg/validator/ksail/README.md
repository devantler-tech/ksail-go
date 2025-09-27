# pkg/validator/ksail

This package provides validation functionality for KSail cluster configurations.

## Purpose

Contains validation logic specifically for KSail cluster configurations (ksail.yaml files). This validator ensures KSail configurations are semantically correct and maintains cross-configuration consistency across different distributions.

## Features

- **KSail Configuration Validation**: Validates KSail cluster configurations for semantic correctness
- **Cross-Configuration Consistency**: Ensures consistency between KSail and distribution-specific configs
- **Multi-Distribution Support**: Validates configurations for Kind, K3d, and EKS distributions
- **Metadata Validation**: Integrates with metadata validator for consistent validation
- **Pre-provisioning Validation**: Catches configuration errors before attempting cluster operations

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator/ksail"

// Create KSail validator
validator := &ksail.Validator{}

// Validate KSail cluster configuration
result := validator.Validate(ksailConfig)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Validation error: %v\n", err)
    }
}
```

## Key Components

- **validator.go**: Core KSail validation logic and implementation
- **cross-configuration validation**: Ensures consistency between ksail.yaml and distribution configs
- **metadata validation**: Leverages shared metadata validation utilities

## Integration

This validator is used by:

- `pkg/config-manager`: KSail configuration loading and validation
- CLI commands that work with ksail.yaml files
- Cluster provisioning workflows

## Validation Rules

The KSail validator enforces:

- Valid KSail cluster naming conventions
- Proper distribution configuration references
- Source directory path validation
- Distribution-specific configuration consistency
- Kubernetes API version and kind validation
- Cross-reference validation between ksail.yaml and distribution configs

## Multi-Distribution Support

The validator handles configurations for:

- **Kind**: Validates consistency with kind.yaml configurations
- **K3d**: Validates consistency with k3d.yaml configurations
- **EKS**: Validates consistency with eksctl.yaml configurations

This ensures KSail configurations are valid and consistent before attempting cluster operations.

---

[⬅️ Go Back](../../README.md)
