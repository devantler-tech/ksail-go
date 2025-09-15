# ui

This directory contains user interface utilities for KSail's CLI.

## Purpose

Provides UI components and utilities for enhancing the command-line user experience. This directory contains packages for visual elements, user notifications, and interface consistency across all KSail CLI commands.

## Features

- **Visual Consistency**: Standardized UI elements across all CLI commands
- **Colored Output**: Terminal colors for better readability
- **Status Symbols**: Visual indicators for different message types
- **ASCII Art**: Visual branding and decorative elements

## Packages

- **[asciiart/](./asciiart/README.md)** - ASCII art and visual elements for CLI branding
- **[notify/](./notify/README.md)** - User notification and messaging utilities with colored output

## Usage

These packages are used internally by CLI commands to provide a consistent and user-friendly interface experience.

```go
import (
    "github.com/devantler-tech/ksail-go/cmd/ui/notify"
    "github.com/devantler-tech/ksail-go/cmd/ui/asciiart"
)

// Provide user feedback
notify.Success("Operation completed successfully")
notify.Error("Operation failed")

// Display branding
asciiart.ShowLogo()
```

---

[⬅️ Go Back](../README.md)