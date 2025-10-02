# Timer Package

The `timer` package provides timing functionality for tracking CLI command execution duration.

## Features

- **Simple API**: Start, stage transitions, and get timing with minimal code
- **Multi-stage support**: Track total time and per-stage time separately
- **Single-stage support**: Simplified timing for operations without stages
- **Zero dependencies**: Uses only Go standard library (`time` package)
- **Testable**: Interface-based design with mockery support

## Installation

This package is part of the KSail project. No separate installation required.

## Usage

### Single-Stage Command

For commands with a single operation phase:

```go
package main

import (
	"fmt"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func main() {
	t := timer.New()
	t.Start()

	// Perform your operation
	doSomething()

	total, _ := t.GetTiming()
	fmt.Printf("Operation completed [%s]\n", total)
}
```

### Multi-Stage Command

For commands with multiple operation phases:

```go
package main

import (
	"fmt"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func main() {
	t := timer.New()
	t.Start()

	// Stage 1
	doStage1()

	t.NewStage("Stage 2")
	// Stage 2
	doStage2()

	t.NewStage("Stage 3")
	// Stage 3
	doStage3()

	total, stage := t.GetTiming()
	fmt.Printf("Operation completed [%s total|%s stage]\n", total, stage)
}
```

### Integration with Notify Package

The timer integrates seamlessly with the `pkg/ui/notify` package:

```go
package main

import (
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func main() {
	t := timer.New()
	t.Start()

	// Perform operation
	err := performOperation()
	if err != nil {
		// No timing on errors
		notify.Error("Operation failed: %v", err)
		return
	}

	// Display timing on success
	total, stage := t.GetTiming()
	timingStr := notify.FormatTiming(total, stage, true) // multi-stage
	notify.Success("Operation completed " + timingStr)
}
```

## API Reference

### Timer Interface

```go
type Timer interface {
	Start()
	NewStage(title string)
	GetTiming() (total, stage time.Duration)
	Stop()
}
```

#### Start()

Initializes timing tracking. Must be called before any other methods. Can be called multiple times to reset the timer.

#### NewStage(title string)

Marks a stage transition. Resets the stage timer while preserving total elapsed time. The title is used for display purposes.

#### GetTiming() (total, stage time.Duration)

Returns current elapsed durations:
- `total`: Time elapsed since Start()
- `stage`: Time elapsed since last NewStage() or Start()

Can be called multiple times without side effects.

#### Stop()

Signals completion of timing. Optional method provided for future extensibility. Currently a no-op.

## Design Principles

1. **Package-First Design**: Timer is a standalone, reusable package
2. **Interface-Based**: Mockable for testing
3. **Single Responsibility**: Only tracks time, doesn't format output
4. **Minimal Overhead**: <1ms overhead per operation
5. **Clean Architecture**: No dependencies on UI or CLI packages

## Testing

The package includes comprehensive contract tests that define the behavioral contract:

```bash
# Run all timer tests
go test ./pkg/ui/timer/

# Run with coverage
go test -cover ./pkg/ui/timer/

# Generate mocks
mockery
```

## Performance

The timer mechanism adds <1ms overhead to command execution. Internal state management uses simple time.Time values and calculations are performed on-demand.

## Contributing

Follow the KSail Constitution principles:
- Package-first design
- Test-first development (TDD)
- Interface-based design
- Clean architecture

See [CONTRIBUTING.md](../../../CONTRIBUTING.md) for details.

## License

Part of the KSail project. See repository root for license information.
