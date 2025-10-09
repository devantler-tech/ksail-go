# pkg/k9s

This package provides a k9s client implementation for KSail.

## Purpose

Provides a wrapper around the k9s terminal UI tool, allowing KSail to launch k9s with appropriate configuration and pass through all k9s flags and subcommands.

## Features

- **k9s Integration**: Launches k9s terminal UI for interactive cluster management
- **Kubeconfig Support**: Automatically configures k9s with the appropriate kubeconfig
- **Flag Pass-through**: All k9s flags and arguments are passed through unchanged
- **Subprocess Management**: Properly manages k9s as a subprocess with stdin/stdout/stderr handling

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/k9s"

// Create a k9s client (uses "k9s" from PATH)
client := k9s.NewClient("")

// Or specify a custom k9s binary path
client := k9s.NewClient("/custom/path/to/k9s")

// Create a connect command that launches k9s
cmd := client.CreateConnectCommand("/path/to/kubeconfig")

// Execute the command (this will launch k9s)
err := cmd.Execute()
if err != nil {
    log.Fatal("Failed to launch k9s:", err)
}
```

## Command Behavior

The `CreateConnectCommand` method creates a Cobra command that:

1. Launches k9s as a subprocess
2. Automatically passes the kubeconfig path via `--kubeconfig` flag
3. Passes through all user-provided arguments and flags to k9s
4. Connects stdin/stdout/stderr so k9s can run interactively

## Flag Pass-through

The connect command uses `DisableFlagParsing: true` to ensure all flags and arguments are passed directly to k9s without being processed by Cobra. This means users can use any k9s flag:

```bash
ksail cluster connect --namespace default
ksail cluster connect --context my-context
ksail cluster connect --readonly
```

## Requirements

This package requires k9s to be installed and available in the system PATH (or specified via a custom path). If k9s is not found, the command will return an error.

## Testing

The package includes comprehensive tests that verify:

- Client creation with custom and default paths
- Command structure and metadata
- Flag pass-through behavior
- Error handling when k9s binary is not found

---

[⬅️ Go Back](../README.md)
