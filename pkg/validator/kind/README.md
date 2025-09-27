# pkg/validator/kind

This package provides validation functionality for Kind cluster configurations.

## Purpose

Contains validation logic specifically for Kind cluster configurations. This validator ensures Kind configurations are valid according to Kind APIs and requirements before cluster provisioning.

## Features

- **Kind Configuration Validation**: Validates Kind cluster configurations using upstream Kind APIs
- **Node Configuration Validation**: Validates node configurations and resource requirements
- **Metadata Validation**: Integrates with metadata validator for consistent validation
- **Port Mapping Validation**: Validates port mappings and network configurations
- **Pre-provisioning Validation**: Catches configuration errors before attempting cluster creation

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator/kind"

// Create Kind validator
validator := kind.NewValidator()

// Validate Kind cluster configuration
result := validator.Validate(kindConfig)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Validation error: %v\n", err)
    }
}
```

## Key Components

- **validator.go**: Core Kind validation logic and implementation
- **Kind API integration**: Uses upstream Kind APIs for validation
- **metadata validation**: Leverages shared metadata validation utilities

## Integration

This validator is used by:

- `pkg/config-manager/kind`: Kind configuration loading and validation
- Kind cluster provisioning workflows
- CLI commands that work with Kind configurations

## Validation Rules

The Kind validator enforces:

- Valid Kind cluster naming conventions
- Proper node configurations and resource limits
- Valid port mapping and network configurations
- Kubernetes version compatibility
- Container image and registry configurations
- Volume mount path validation
- CNI and networking plugin configurations

This ensures Kind configurations are valid before attempting local cluster provisioning with Docker.

---

[⬅️ Go Back](../../README.md)
