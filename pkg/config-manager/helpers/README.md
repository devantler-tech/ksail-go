# pkg/config-manager/helpers

This package provides common functionality for configuration managers to eliminate duplication across different distributions.

## Purpose

Contains shared utilities and helper functions used by various configuration managers (Kind, K3d, EKS) to avoid code duplication. This package provides the foundational configuration loading, validation, and management functionality.

## Features

- **Configuration Loading**: Generic configuration loading from YAML files
- **Validation Integration**: Seamless integration with the validator package
- **Error Handling**: Standardized error types and handling patterns
- **File I/O Operations**: Safe file operations with proper error handling
- **Marshalling Support**: Integration with YAML marshalling utilities
- **Path Resolution**: Robust path handling and resolution

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"

// Load and validate configuration from file
config, err := helpers.LoadConfig[MyConfigType](
    "/path/to/config.yaml",
    validator.ValidateFunc,
)
if err != nil {
    // Handle validation or loading errors
}
```

## Key Components

- **loader.go**: Core configuration loading and validation logic
- **common functionality**: Shared utilities across all config managers
- **error definitions**: Standardized error types for configuration operations
- **testutils/**: Testing utilities specific to config-manager helpers

## Integration

This package is used by:

- `pkg/config-manager/kind`: Kind cluster configuration management
- `pkg/config-manager/k3d`: K3d cluster configuration management
- `pkg/config-manager/eks`: EKS cluster configuration management

This design ensures consistent behavior and reduces code duplication across all supported Kubernetes distributions.

---

[⬅️ Go Back](../README.md)
