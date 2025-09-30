# pkg/validator/testutils

This package provides common test utilities for validator tests to eliminate duplication across validator packages.

## Purpose

Contains shared testing utilities, common test case patterns, and helper functions specifically designed for testing validator packages. This package provides standardized testing patterns to ensure consistent validation testing across all validator types.

## Features

- **Common Test Cases**: Standardized test case structures for validator testing
- **Test Utilities**: Helper functions for setting up and running validator tests
- **Error Testing**: Patterns for testing validation error scenarios
- **Result Validation**: Utilities for validating ValidationResult outputs
- **Generic Test Support**: Type-safe test utilities using Go generics

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/validator/testutils"

func TestValidator(t *testing.T) {
    testCases := []testutils.ValidatorTestCase[MyConfigType]{
        {
            Name:          "valid config",
            Config:        validConfig,
            ExpectedValid: true,
        },
        {
            Name:          "invalid config",
            Config:        invalidConfig,
            ExpectedValid: false,
        },
    }

    validator := &MyValidator{}
    testutils.RunValidatorTests(t, validator, testCases)
}
```

## Key Components

- **common.go**: Common test utilities and test case structures
- **ValidatorTestCase**: Generic test case structure for validator testing
- **test helpers**: Utility functions for running standardized validator tests

## Integration

This package is used by all validator packages for testing:

- `pkg/validator/kind`: Kind validator testing
- `pkg/validator/k3d`: K3d validator testing
- `pkg/validator/ksail`: KSail validator testing
- `pkg/validator/metadata`: Metadata validator testing

## Test Patterns

The package provides standardized patterns for:

- **Valid Configuration Testing**: Ensuring valid configs pass validation
- **Invalid Configuration Testing**: Ensuring invalid configs fail validation
- **Error Message Testing**: Validating specific error messages and types
- **Edge Case Testing**: Common edge cases across all validator types

This ensures consistent and comprehensive testing across all validator functionality in KSail.

---

[⬅️ Go Back](../../README.md)
