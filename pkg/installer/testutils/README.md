# pkg/installer/testutils

This package provides testing utilities for installer package testing.

## Purpose

Contains testing utilities and helper functions specifically designed for testing installer packages (kubectl, flux, etc.). This package provides kubeconfig management, temporary file utilities, and testing patterns for installer functionality.

## Features

- **Kubeconfig Testing**: Helper functions for managing kubeconfig files in tests
- **Temporary File Management**: Utilities for creating temporary test files with proper permissions
- **Installation Testing**: Common patterns for testing installer functionality
- **File Permission Testing**: Utilities for testing file permission scenarios
- **Test Constants**: Standard permissions and testing configurations

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/installer/testutils"

func TestInstaller(t *testing.T) {
    // Create temporary kubeconfig for testing
    kubeconfig := testutils.CreateTempKubeconfig(t, kubeconfigContent)
    defer os.Remove(kubeconfig)

    // Test installer functionality with temporary files
}
```

## Key Components

- **kubeconfig_helpers.go**: Kubeconfig testing utilities and helper functions
- **file permissions**: Standard file permission constants for testing
- **temporary file management**: Helper functions for test file creation

## Integration

This package is used by installer packages for testing:

- `pkg/installer/kubectl`: kubectl installation testing
- `pkg/installer/flux`: Flux installation testing

This ensures consistent testing patterns across all installer functionality.

---

[⬅️ Go Back](../../README.md)
