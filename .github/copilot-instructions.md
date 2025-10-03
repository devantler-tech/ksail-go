# KSail-Go

KSail is a Go-based CLI tool for managing Kubernetes clusters and workloads. It provides declarative cluster provisioning, workload management, and lifecycle operations for Kind, K3d, and EKS distributions.

**ALWAYS reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.**

## Working Effectively

### Prerequisites

Install these exact tools before starting development:

- Go (v1.23.9+) - `go version` should show go1.23.9 or higher
- mockery (v3.x) - for generating test mocks
- golangci-lint - for code linting
- mega-linter-runner - for comprehensive validation
- Docker - required for cluster provisioning and system tests

### Bootstrap and Build Commands

Run these commands in sequence to set up the development environment:

```bash
# Download dependencies (very fast if cached)
go mod download

# Build all packages - takes ~2 seconds
go build ./...

# Build the main CLI binary - takes ~1.4 seconds
go build -o ksail .

# Generate mocks - takes ~1.2 seconds
mockery

# Run unit tests - takes ~37 seconds. NEVER CANCEL - set timeout to 60+ seconds
go test ./...

# Run linter - takes ~1m16s. NEVER CANCEL - set timeout to 90+ seconds
golangci-lint run
```

### Mega-Linter (Comprehensive Validation)

```bash
# Run comprehensive linting with go flavor - takes 5+ minutes
# NEVER CANCEL: This is thorough validation. Set timeout to 10+ minutes
mega-linter-runner -f go
```

## Validation

### Always Run Before Committing

Execute these commands before any commit to ensure CI will pass:

```bash
# Essential pre-commit validation (run all of these):
mockery                    # Generate fresh mocks
go test ./...             # Run all tests (~37s)
golangci-lint run         # Lint code (~1m16s)
go build -o ksail .       # Ensure clean build
```

### Manual Testing Scenarios

**ALWAYS test actual CLI functionality after making changes by running these scenarios:**

#### Basic CLI Validation

```bash
# Test CLI help and version
./ksail --help
./ksail --version

# Test all command help outputs
./ksail init --help
./ksail up --help
./ksail down --help
./ksail status --help
```

#### Complete Cluster Lifecycle Test

Run this complete scenario in a temporary directory to validate changes:

```bash
# Create test directory and navigate to it
mkdir -p /tmp/ksail-test && cd /tmp/ksail-test

# Test Kind distribution (most common)
./ksail init --distribution Kind
./ksail up
./ksail status
./ksail list
./ksail start
./ksail reconcile
./ksail down

# Clean up test files
rm -rf k8s kind.yaml ksail.yaml
```

#### Alternative Distribution Testing

```bash
# Test K3d distribution
./ksail init --distribution K3d

# Test EKS distribution
./ksail init --distribution EKS
```

### System Tests

The CI runs comprehensive system tests that validate:

- `init --distribution Kind`
- `init --distribution K3d`
- `init --distribution EKS`

Each runs the complete lifecycle: init → up → status → list → start → reconcile → down

## Project Structure and Navigation

### Repository Layout

```txt
/home/runner/work/ksail-go/ksail-go/
├── cmd/                    # CLI commands using Cobra framework
│   ├── *.go               # Command implementations (init.go, up.go, down.go, etc.)
│   ├── ui/                # User interface utilities
│   └── internal/          # Command helper utilities
├── pkg/                   # Core business logic packages
│   ├── apis/              # Kubernetes API definitions
│   ├── config-manager/    # Configuration management
│   ├── installer/         # Component installation utilities
│   ├── io/                # Safe file I/O operations
│   ├── provisioner/       # Cluster provisioning and lifecycle
│   ├── scaffolder/        # Project scaffolding
│   └── validator/         # Validation utilities
├── internal/              # Internal utility packages
│   └── utils/             # Common utilities (k8s, path)
├── main.go               # Application entry point
├── go.mod                # Go module definition
├── .github/workflows/    # CI/CD pipeline definitions
└── scripts/              # Build and development scripts
```

### Key Files to Review When Making Changes

- **Command changes**: Always check corresponding test files (`*_test.go`)
- **API changes**: Review `pkg/apis/cluster/v1alpha1/` definitions
- **Configuration changes**: Check `pkg/config-manager/` implementations
- **Provisioner changes**: Validate against `pkg/provisioner/cluster/` providers
- **Build/CI changes**: Review `.github/workflows/ci.yaml`

### Common Development Locations

- **Adding new CLI commands**: `cmd/*.go` + corresponding tests
- **Cluster provider logic**: `pkg/provisioner/cluster/{kind,k3d,eks}/`
- **Configuration handling**: `pkg/config-manager/`
- **File generation**: `pkg/io/generator/`
- **Test utilities**: `internal/testutils/` and package-specific `testutils/`

## Timing Expectations and Timeouts

### Command Timing Reference (based on validation)

- `go mod download`: ~0.045s (when cached)
- `go build ./...`: ~2.1s
- `go build -o ksail .`: ~1.4s
- `go test ./...`: ~37s (full test suite)
- `golangci-lint run`: ~1m16s
- `mockery`: ~1.2s
- `mega-linter-runner -f go`: 5+ minutes

### Recommended Timeout Settings

> [!CAUTION]
> CRITICAL: NEVER CANCEL these operations prematurely

- Build commands: 60+ seconds timeout
- Test commands: 90+ seconds timeout
- Linter commands: 120+ seconds timeout
- Mega-linter: 600+ seconds (10+ minutes) timeout

## CI Workflow Information

### GitHub Actions Pipeline

The CI pipeline (`.github/workflows/ci.yaml`) runs:

1. **Standard Go CI**: Build, test, lint using reusable workflows
2. **System Tests**: Matrix testing across Kind, K3d, and EKS distributions
3. **Full lifecycle validation**: Each distribution tested through complete workflow

### Pre-commit Hooks

Pre-commit hooks automatically run:

- `golangci-lint-fmt`: Go code formatting
- `mockery`: Mock generation via `scripts/run-mockery.sh`

Install pre-commit hooks: `pre-commit install`

## Dependencies and Requirements

### Go Version

- **Required**: Go 1.24.0+ (specified in go.mod)
- **Validated**: Works with Go 1.25.1

### External Tools

- **Docker**: Required for cluster provisioning (Kind, K3d containers)
- **mockery v3.x**: Critical for test mock generation
- **golangci-lint**: Code quality enforcement
- **mega-linter**: Comprehensive project validation

### Key Go Dependencies

- `github.com/spf13/cobra`: CLI framework
- `sigs.k8s.io/kind`: Kind cluster management
- `github.com/k3d-io/k3d/v5`: K3d cluster management
- `github.com/weaveworks/eksctl`: EKS cluster management
- `k8s.io/client-go`: Kubernetes client libraries

## Common Tasks Reference

### Building the Application

```bash
# Development build
go build -o ksail .

# Cross-platform build (example)
GOOS=linux GOARCH=amd64 go build -o ksail-linux .
GOOS=darwin GOARCH=amd64 go build -o ksail-darwin .
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package tests
go test ./cmd
go test ./pkg/provisioner/cluster/kind

# Verbose test output
go test -v ./cmd

# Test with coverage
go test -cover ./...
```

### Mock Management

```bash
# Generate all mocks (uses .mockery.yml config)
mockery

# Check mockery configuration
mockery showconfig
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Fix auto-fixable issues
golangci-lint run --fix

# Comprehensive validation
mega-linter-runner -f go
```

## CLI Command Timing Feature

### Overview

All KSail CLI commands display timing information on successful completion to help users monitor performance of cluster operations.

### Timer Package (`pkg/ui/timer`)

**Location**: `pkg/ui/timer/`
**Purpose**: Provides timing tracking functionality for CLI command execution.

#### Usage Pattern (Single-Stage Command)

```go
package cmd

import (
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
)

func HandleCommandRunE(cmd *cobra.Command, ...) error {
	// Create and start timer
	tmr := timer.New()
	tmr.Start()

	// Execute command logic
	err := doSomething()
	if err != nil {
		// NO timing on error paths
		return fmt.Errorf("operation failed: %w", err)
	}

	// Get timing and format (false = single-stage)
	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, false)

	// Display success with timing
	notify.Successf(cmd.OutOrStdout(), "operation complete %s", timingStr)
	return nil
}
```

#### Usage Pattern (Multi-Stage Command)

```go
func HandleMultiStageCommandRunE(cmd *cobra.Command, ...) error {
	// Create and start timer
	tmr := timer.New()
	tmr.Start()

	// Stage 1
	notify.Titleln(cmd.OutOrStdout(), "🚀", "Starting...")
	err := doStage1()
	if err != nil {
		return fmt.Errorf("stage 1 failed: %w", err)
	}

	// Transition to stage 2
	tmr.NewStage()
	notify.Titleln(cmd.OutOrStdout(), "📦", "Deploying...")
	err = doStage2()
	if err != nil {
		return fmt.Errorf("stage 2 failed: %w", err)
	}

	// Get timing and format (true = multi-stage)
	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, true)

	notify.Successf(cmd.OutOrStdout(), "operation complete %s", timingStr)
	return nil
}
```

### Timing Display Formats

- **Single-stage**: `[1.2s]`
- **Multi-stage**: `[5m30s total|2m15s stage]`
- **Sub-second**: `[500ms]` or `[123µs]`
- **Long durations**: `[1h23m45s total|15m0s stage]`

### Constitutional Compliance

- ✅ **Package-First Design**: Timer is a standalone `pkg/ui/timer` package
- ✅ **Test-First Development**: All contract tests written before implementation
- ✅ **Interface-Based**: Timer interface with mockery support
- ✅ **<1ms Overhead**: Timer adds negligible performance impact
- ✅ **Clean Architecture**: Timer has no dependency on notify (one-way integration)

### Testing Timer Integration

```bash
# Run timer package tests
go test ./pkg/ui/timer/... -v

# Run notify format timing tests
go test ./pkg/ui/notify/... -run FormatTiming -v

# Test CLI command with timing
./ksail init --distribution Kind
# Expected output: "✔ initialized project [1.2s]"
```
