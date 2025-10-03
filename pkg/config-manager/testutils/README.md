# pkg/config-manager/testutils

This package provides testing utilities for config manager packages.

## Purpose

Offers shared test helpers, scenario definitions, and reusable patterns for
validating configuration managers across distributions.

## Features

- **Test Scenario Utilities**: Generic helpers for defining config manager test cases
- **Temporary File Management**: Helpers for working with temporary config files in tests
- **Caching Validation**: Utilities to verify configuration caching behavior
- **Error Scenario Support**: Patterns for asserting configuration load failures

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/config-manager/testutils"

func TestConfigManager(t *testing.T) {
    scenarios := []testutils.TestScenario[MyConfig]{
        {
            Name:           "valid config",
            ConfigContent:  "...",
            UseCustomConfigPath: true,
            ValidationFunc: func(t *testing.T, config *MyConfig) {
                // add assertions
            },
        },
    }

    testutils.RunConfigManagerTests(t, newManager, scenarios)
}
```

## Key Components

- **helpers.go**: Core helper functions, scenario definitions, and caching tests
- **test constants**: Shared permissions and path helpers for test fixtures

This package ensures consistent and reliable testing patterns for configuration
management functionality.

---

[⬅️ Go Back](../README.md)
