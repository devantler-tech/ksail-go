# pkg/config-manager

This package provides centralized configuration management using Viper for KSail.

## Purpose

Provides a generic, type-safe configuration management interface that handles:

- Loading configuration from files and environment variables
- Providing access to the underlying Viper instance for flag binding
- Type-safe configuration handling through Go generics

## Features

- **Generic Interface**: `ConfigManager[T any]` supports any configuration type
- **Viper Integration**: Uses Viper for configuration loading and management
- **Environment Variable Support**: Automatically loads from environment variables
- **Flag Binding**: Provides access to Viper instance for CLI flag binding

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/config-manager"

// Create a config manager for your config type
var manager ConfigManager[MyConfigType]

// Load configuration
config, err := manager.LoadConfig()
if err != nil {
    log.Fatal(err)
}

// Get Viper instance for flag binding
viper := manager.GetViper()
```

## Implementation

The KSail-specific configuration implementation lives in `pkg/config-manager/ksail`, providing the concrete manager used by the CLI. This package continues to define the generic interface.

## Related Packages

- **[pkg/config-manager/ksail/](./ksail/README.md)** - KSail-specific configuration implementation

---

[⬅️ Go Back](../README.md)
