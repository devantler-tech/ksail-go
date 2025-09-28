# cmd

This package contains the command-line interface (CLI) implementation for KSail.

## Purpose

Implements the CLI commands and user interface for KSail using the Cobra framework. This package provides the main entry point for all KSail operations through command-line interactions.

## Structure

- **Root Command**: `root.go` - Main CLI setup with Cobra framework and version handling
- **Command Implementations**: Individual command files (`init.go`, `cluster/`, etc.)
- **UI Components**: `ui/` subdirectory containing user interface utilities
- **Internal Helpers**: `internal/` subdirectory containing command helper utilities

## Available Commands

- `init` - Initialize a new KSail project
- `cluster` - Parent namespace for cluster lifecycle commands (`up`, `down`, `start`, `stop`, `status`, `list`)
- `workload` - Namespace for workload-focused operations (`reconcile`, `apply`, `install` placeholders)

## Features

- **Cobra Framework**: Uses Cobra for consistent CLI structure and help generation
- **Colored Output**: Colored terminal output for better user experience
- **Consistent UI**: Standardized symbols and messaging across all commands
- **Help System**: Comprehensive help and usage information for all commands

## Subpackages

- `ui/asciiart/` - ASCII art and visual elements
- `ui/notify/` - User notification and messaging utilities
- `internal/cmdhelpers/` - Internal command helper utilities

## Usage

The CLI is built and used as:

```bash
# Build the CLI
go build -o ksail .

# Use the CLI
./ksail --help
./ksail init --distribution Kind
./ksail cluster up
./ksail cluster status
./ksail cluster down
./ksail workload --help
./ksail workload reconcile
```

This package serves as the primary user interface for KSail, providing a comprehensive command-line experience for managing Kubernetes clusters and workloads.

---

[⬅️ Go Back](../README.md)
