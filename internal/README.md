# internal

This directory contains internal utility packages for KSail.

## Purpose

Houses internal packages that provide common functionality shared across different KSail components. These packages are implementation details and are not part of KSail's public API.

## Features

- **Internal Use Only**: Not intended for external consumption
- **Shared Utilities**: Common functionality used across multiple KSail components
- **Implementation Details**: Supporting code that may change without notice

## Packages

- **[internal/utils/](./utils/README.md)** - General utility functions and helpers

## Usage

These packages are for internal use within KSail components only:

```go
import "github.com/devantler-tech/ksail-go/internal/utils/path"

// Internal usage only - API may change without notice
```

**Note**: Packages in `internal/` are not part of KSail's public API and may change without notice. External applications should not depend on these packages directly.

---

[⬅️ Go Back](../README.md)
