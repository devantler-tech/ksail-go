# ui/asciiart

This package provides ASCII art functionality for KSail's CLI interface.

## Purpose

Contains ASCII art generators and utilities for enhancing the visual presentation of KSail's command-line interface. This package provides branded ASCII art and visual elements to improve the user experience of the CLI.

## Features

- **ASCII Art Generation**: Creates ASCII art for logos, banners, and decorative elements
- **CLI Enhancement**: Improves the visual appeal of command-line output
- **Branding**: Provides consistent visual branding for KSail CLI
- **Customizable Output**: Configurable ASCII art for different contexts

## Usage

```go
import "github.com/devantler-tech/ksail-go/cmd/ui/asciiart"

// Display ASCII art in CLI
asciiart.ShowLogo()
asciiart.ShowBanner("Welcome to KSail")
```

This package is used by KSail's CLI commands to provide a more engaging and visually appealing user interface, particularly during application startup and key operations.

---

[⬅️ Go Back](../README.md)