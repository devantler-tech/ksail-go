# pkg/cliwrapper

This package provides utilities for wrapping urfave/cli v1 applications within Cobra-based CLIs.

## Purpose

Some Go applications use the `urfave/cli` framework instead of Cobra. This package provides a generic way to wrap such applications so they can be embedded as subcommands in a Cobra-based CLI.

## Usage

### Basic Wrapping

```go
import (
    "github.com/devantler-tech/ksail-go/pkg/cliwrapper"
    "github.com/urfave/cli"
)

// Create a urfave/cli app
app := cli.NewApp()
app.Name = "myapp"
app.Usage = "My application"
app.Action = func(c *cli.Context) error {
    // ... app logic
    return nil
}

// Wrap it in a Cobra command
cobraCmd := cliwrapper.WrapCliApp(app)

// Add to your Cobra root command
rootCmd.AddCommand(cobraCmd)
```

### With Custom IO

```go
cobraCmd := cliwrapper.WrapCliAppWithIO(app, stdin, stdout, stderr)
```

### For Testing

```go
stdout, stderr, err := cliwrapper.CaptureCliAppOutput(app, []string{"arg1", "arg2"})
```

## Limitations

- **Flag Parsing**: Uses `DisableFlagParsing` to pass all flags to the urfave/cli app
- **Help Integration**: urfave/cli help is shown instead of Cobra help
- **IO Redirection**: May have limitations with certain IO patterns

## When to Use

Use this wrapper when:
- You need to embed a urfave/cli-based tool in your Cobra CLI
- The tool doesn't provide a Cobra-compatible API
- Creating a native integration would require significant code duplication

For SOPS specifically, we use exec instead because:
1. SOPS doesn't export an app builder function
2. All commands are defined in a 2500+ line main.go file
3. Wrapping would require copying significant code and would break on updates

## Related

- `pkg/sops` - SOPS integration using exec (cleaner for this specific use case)
- `pkg/kubectl` - kubectl integration using native Cobra commands (ideal when available)
