# pkg/validator

This package provides interfaces and types for configuration file validation across KSail.

## Purpose

Contains the core validation interfaces, types, and shared functionality used by all validator packages. This package defines the validation contract that all specific validators (Kind, K3d, EKS, KSail, metadata) must implement.

## Features

- **Validation Interfaces**: Generic interfaces for configuration validation
- **Validation Types**: Common types for validation results and error handling
- **Thread-Safe Design**: All validators are designed for concurrent use
- **Generic Type Support**: Type-safe validation using Go generics
- **Extensible Architecture**: Easy to add new validator types

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator"

// Implement the Validator interface
type MyValidator struct{}

func (v *MyValidator) Validate(config MyConfigType) validator.ValidationResult {
    // Perform semantic validation
    return validator.ValidationResult{
        Valid:  true,
        Errors: nil,
    }
}

// Use a validator
result := validator.Validate(config)
if !result.Valid {
    // Handle validation errors
}
```

## Key Components

- **interfaces.go**: Core validation interfaces and contracts
- **types.go**: Validation result types and error handling
- **thread-safe design**: All implementations suitable for concurrent use

## Sub-packages

This package provides validation for different configuration types:

- `eks/`: EKS cluster configuration validation
- `k3d/`: K3d cluster configuration validation
- `kind/`: Kind cluster configuration validation
- `ksail/`: KSail project configuration validation
- `metadata/`: Metadata validation utilities
- `testutils/`: Testing utilities for validator packages

## Architecture

The validator package follows a generic, interface-based design that ensures:

- **Type Safety**: Generic interfaces provide compile-time type checking
- **Consistency**: All validators follow the same validation contract
- **Extensibility**: Easy to add new validator types
- **Testability**: Interfaces enable easy mocking and testing

---

[⬅️ Go Back](../../README.md)
