# cmd/internal

This directory contains internal packages for KSail's CLI commands.

## Purpose

Houses internal utilities and helper packages that are used by KSail's CLI command implementations. These packages provide common functionality that is shared across different commands to ensure consistency and reduce code duplication.

## Features

- **Command Helpers**: Utilities for common CLI command patterns
- **Internal Use**: Not intended for external consumption
- **Consistency**: Ensures consistent behavior across all CLI commands
- **Code Reuse**: Reduces duplication in command implementations

## Packages

- **[cmd/internal/helpers/](./helpers/command.go)** - Helper utilities for CLI command implementation
- **testutils/** - Testing utilities for CLI commands

## Usage

These packages are for internal use within KSail's CLI implementation:

```go
import helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"

// Internal usage within CLI commands
cmd := helpers.NewCobraCommand(...)
```

**Note**: Packages in `cmd/internal/` are implementation details and may change without notice. External applications should not depend on these packages directly.

---

[⬅️ Go Back](../README.md)
