# pkg/validator/k3d

This package provides validation functionality for K3d cluster configurations.

## Purpose

Contains validation logic specifically for K3d cluster configurations. This validator ensures K3d configurations are valid according to K3d APIs and requirements before cluster provisioning.

## Features

- **K3d Configuration Validation**: Validates K3d cluster configurations using upstream K3d APIs
- **Runtime Validation**: Validates container runtime compatibility and requirements
- **Metadata Validation**: Integrates with metadata validator for consistent validation
- **Port Mapping Validation**: Validates port mappings and network configurations
- **Pre-provisioning Validation**: Catches configuration errors before attempting cluster creation

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator/k3d"

// Create K3d validator
validator := &k3d.Validator{}

// Validate K3d cluster configuration
result := validator.Validate(k3dConfig)
if !result.Valid {
    for _, err := range result.Errors {
        fmt.Printf("Validation error: %v\n", err)
    }
}
```

## Key Components

- **validator.go**: Core K3d validation logic and implementation
- **K3d API integration**: Uses upstream K3d APIs for validation
- **runtime validation**: Validates container runtime compatibility
- **metadata validation**: Leverages shared metadata validation utilities

## Integration

This validator is used by:

- `pkg/config-manager/k3d`: K3d configuration loading and validation
- K3d cluster provisioning workflows
- CLI commands that work with K3d configurations

## Validation Rules

The K3d validator enforces:

- Valid K3d cluster naming conventions
- Proper container runtime configurations (Docker/Podman)
- Valid port mapping and network configurations
- Kubernetes version compatibility with K3s
- Volume mount path validation
- Registry and image configuration validation
- Resource limits and node configurations

This ensures K3d configurations are valid before attempting local cluster provisioning.

---

[⬅️ Go Back](../../README.md)
