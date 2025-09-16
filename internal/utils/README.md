# internal/utils

This directory contains internal utility packages for KSail.

## Purpose

Provides common utility functions and helpers that are used across different KSail components. These utilities handle cross-cutting concerns like path operations and Kubernetes-related helper functions.

## Features

- **Cross-Platform Utilities**: Functions that work consistently across different operating systems
- **Kubernetes Helpers**: Utilities for working with Kubernetes APIs and resources
- **Path Operations**: Safe and secure file system path handling
- **Internal Use**: Designed for internal KSail usage only

## Packages

- **[internal/utils/k8s/](./k8s/README.md)** - Kubernetes utilities and helper functions
- **[internal/utils/path/](./path/README.md)** - Path utilities for file system operations

## Usage

These are internal utility packages for use within KSail:

```go
import (
    "github.com/devantler-tech/ksail-go/internal/utils/k8s"
    "github.com/devantler-tech/ksail-go/internal/utils/path"
)

// Internal usage within KSail components
```

**Note**: These are internal packages and their APIs may change without notice. External applications should not depend on these packages directly.

---

[⬅️ Go Back](../README.md)
