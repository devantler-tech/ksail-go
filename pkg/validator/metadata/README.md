# pkg/validator/metadata

This package provides shared metadata validation utilities used across multiple validators.

## Purpose

Contains common metadata validation functions that are shared across different validator packages. This package provides consistent validation of Kubernetes-style metadata fields like `kind`, `apiVersion`, and other common configuration fields.

## Features

- **Metadata Field Validation**: Validates common metadata fields (kind, apiVersion)
- **Consistent Error Reporting**: Standardized error messages and validation results
- **Shared Validation Logic**: Eliminates duplication across validator packages
- **Extensible Design**: Easy to add new metadata validation rules
- **Integration Ready**: Designed for use by all validator types

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator/metadata"

// Validate metadata fields
result := &validator.ValidationResult{}
metadata.ValidateMetadata(
    config.Kind,
    config.APIVersion,
    "Cluster",
    "v1alpha1",
    result,
)

if !result.Valid {
    // Handle validation errors
}
```

## Key Components

- **metadata.go**: Core metadata validation functions and utilities
- **field validation**: Common validation patterns for metadata fields
- **error integration**: Seamless integration with validator result types

## Integration

This package is used by all specific validator packages:

- `pkg/validator/kind`: Kind configuration metadata validation
- `pkg/validator/k3d`: K3d configuration metadata validation
- `pkg/validator/eks`: EKS configuration metadata validation
- `pkg/validator/ksail`: KSail configuration metadata validation

## Validation Functions

The package provides validation for:

- **Kind Field**: Validates required kind field and expected values
- **API Version**: Validates required apiVersion field and compatibility
- **Field Presence**: Ensures required metadata fields are present
- **Value Matching**: Validates metadata values match expected patterns

This ensures consistent metadata validation across all configuration types in KSail.

---

[⬅️ Go Back](../../README.md)
