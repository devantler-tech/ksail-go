# cmd/ui/notify

This package provides notification utilities for KSail's CLI interface.

## Purpose

Contains utilities for sending user notifications through the command-line interface. This package provides consistent messaging, status updates, and user feedback throughout KSail's CLI operations.

## Features

- **Colored Output**: Uses `fatih/color` for colored terminal output
- **Status Symbols**: Provides consistent symbols for different message types:
  - `✔` - Success messages (SuccessSymbol)
  - `✗` - Error messages (ErrorSymbol)  
  - `⚠` - Warning messages (WarningSymbol)
  - `►` - Activity messages (ActivitySymbol)
- **Consistent Messaging**: Standardized format for CLI notifications
- **Writer Abstraction**: Supports custom output writers for testing

## Usage

```go
import "github.com/devantler-tech/ksail-go/cmd/ui/notify"

// Send different types of notifications
notify.Success("Cluster created successfully")
notify.Error("Failed to create cluster")
notify.Warning("Cluster already exists")
notify.Activity("Creating cluster...")
```

This package ensures consistent and visually clear communication between KSail and users through the command-line interface, improving the overall user experience.

---

[⬅️ Go Back](../README.md)
