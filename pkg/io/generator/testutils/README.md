# pkg/io/generator/testutils

This package provides testing utilities specifically for generator package testing.

## Purpose

Contains testing utilities and helper functions designed for testing generator packages. This package provides marshal failure testing, generic test utilities, and common testing functionality for file generation and I/O operations.

## Features

- **Marshal Failure Testing**: Utilities for testing marshal/unmarshal error scenarios
- **Generator Testing**: Common patterns for testing file generation functionality
- **Common Test Utilities**: Shared testing functionality across generator packages
- **Error Scenario Testing**: Utilities for testing various error conditions
- **File Generation Testing**: Patterns for testing file creation and content validation

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"

func TestGenerator(t *testing.T) {
    // Use marshal failure testing utilities
    failer := testutils.NewMarshalFailer()

    // Test error scenarios in generators
    err := generator.Generate(failer)
    assert.Error(t, err, "expected marshal failure")
}
```

## Key Components

- **generator.go**: Core generator testing utilities
- **marshal_failer.go**: Marshal failure testing utilities and error simulation
- **common_tests.go**: Common test patterns shared across generator packages
- **doc.go**: Package documentation and overview

## Integration

This package is used by generator packages for testing:

- `pkg/io/generator/kind`: Kind configuration generation testing
- `pkg/io/generator/k3d`: K3d configuration generation testing
- `pkg/io/generator/yaml`: YAML generation testing
- `pkg/io/generator/kustomization`: Kustomization file generation testing

This ensures consistent testing patterns across all generator functionality.

---

[⬅️ Go Back](../../../README.md)
