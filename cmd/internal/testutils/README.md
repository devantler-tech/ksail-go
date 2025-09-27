# cmd/internal/testutils

This package provides testing utilities specifically for command-layer testing in KSail.

## Purpose

Contains command-specific testing helpers, utilities, and shared test errors designed for testing CLI commands and command-layer functionality. This package provides command-layer specific testing functionality that complements the general-purpose testing utilities.

## Features

- **Command Testing Helpers**: Utilities for testing CLI commands and interactions
- **Mock Configuration**: Helper functions for creating test configuration managers
- **Error Handling**: Shared test errors for consistent command testing
- **Cobra Integration**: Testing utilities for Cobra CLI commands
- **Snapshot Testing**: Integration with snapshot testing for command outputs

## Usage

```go
import "github.com/devantler-tech/ksail-go/cmd/internal/testutils"

// Create a default config manager for testing
configManager := testutils.CreateDefaultConfigManager()

// Use in command tests for consistent testing setup
func TestCommand(t *testing.T) {
    // Command-specific test utilities
}
```

## Key Components

- **helpers.go**: Core testing helper functions for command testing
- **errors.go**: Shared test errors for command validation
- **doc.go**: Package documentation and overview

This package ensures consistent and reliable testing patterns across all KSail CLI commands.

---

[⬅️ Go Back](../../README.md)
