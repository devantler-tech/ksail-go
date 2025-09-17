# cmd/config-manager

This package provides KSail-specific configuration management implementation.

## Purpose

Implements the `ConfigManager` interface specifically for KSail configuration, handling the loading and management of KSail-specific settings and preferences. This package was moved from `pkg/config-manager/ksail` to avoid circular dependencies with `cmd/ui/notify`.

## Usage

```go
import configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"

// Use KSail-specific configuration manager
manager := configmanager.NewConfigManager()
config, err := manager.LoadConfig()
if err != nil {
    log.Fatal(err)
}
```

This package contains the concrete implementation of configuration management tailored for KSail's specific configuration structure and requirements.

---

[⬅️ Go Back](../README.md)
