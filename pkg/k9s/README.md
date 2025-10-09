# pkg/k9s

This package provides a k9s client implementation for KSail.

## Purpose

Provides a wrapper around the k9s terminal UI tool using its native Go packages, allowing KSail to launch k9s with appropriate configuration.

## Features

- **Native k9s Integration**: Uses k9s's Go packages directly (no subprocess execution)
- **Kubeconfig Support**: Automatically configures k9s with the appropriate kubeconfig
- **Argument Pass-through**: All k9s arguments are passed through to the native k9s execution

## Usage

```go
import "github.com/devantler-tech/ksail-go/pkg/k9s"

// Create a k9s client
client := k9s.NewClient()

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

1. Launches k9s using its native Go packages
2. Automatically passes the kubeconfig path via `--kubeconfig` flag
3. Passes through all user-provided arguments to k9s
4. Connects k9s to the terminal for interactive use

## Native Integration

This package integrates k9s using its exported `cmd.Execute()` function from `github.com/derailed/k9s/cmd`. This provides:

- Direct integration without subprocess overhead
- Full access to k9s's native flags and features
- Consistent behavior with standalone k9s

## Testing

The package includes comprehensive tests that verify:

- Client creation
- Command structure and metadata
- Proper kubeconfig handling

---

[⬅️ Go Back](../README.md)
