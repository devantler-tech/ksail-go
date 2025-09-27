# pkg/config-manager/helpers/testutils

This package provides testing utilities specifically for config-manager helper testing.

## Purpose

Contains testing utilities and helper functions designed specifically for testing the config-manager helpers package. This package provides test suite utilities, temporary file management, and testing patterns for configuration management testing.

## Features

- **Test Suite Utilities**: Common test patterns for config-manager testing
- **Temporary File Management**: Helper functions for creating test configuration files
- **Configuration Testing**: Utilities for testing configuration loading and validation
- **Error Testing**: Patterns for testing configuration error scenarios
- **File Permission Testing**: Utilities for testing file permission scenarios

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/config-manager/helpers/testutils"

func TestConfigManager(t *testing.T) {
    // Use test utilities for config manager testing
    suite := testutils.NewTestSuite(t)

    // Create temporary configuration files
    configFile := suite.CreateTempConfigFile("test-config.yaml", configContent)
    defer os.Remove(configFile)
}
```

## Key Components

- **suite.go**: Test suite utilities and common testing patterns
- **temporary file management**: Helper functions for test file creation
- **test constants**: Standard permissions and testing configurations

This package ensures consistent and reliable testing patterns for configuration management functionality.

---

[⬅️ Go Back](../../README.md)
