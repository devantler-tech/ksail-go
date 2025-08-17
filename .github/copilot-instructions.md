# KSail Go - Kubernetes Cluster Management CLI

KSail is a Go-based CLI tool for managing local Kubernetes clusters and workloads declaratively. **Currently a work-in-progress migration from a previous implementation.** The core CLI structure exists but most functionality is planned/under development.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build

- **Install Go 1.24.6+**: Verify with `go version`
- **Download dependencies**: `go mod download` (completes in ~5 seconds)
- **Build the application**: `go build -o ksail .` -- takes <0.1 seconds when dependencies cached. Set timeout to 60+ seconds for safety.
- **Install golangci-lint v2**: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/go/bin latest`

### Required Dependencies

- **Go 1.24.6+**: Programming language runtime (required)
- **Docker**: Container runtime - verify with `docker --version` (for future functionality)
- **Kind**: Local Kubernetes clusters - verify with `kind version` (for future functionality)  
- **kubectl**: Kubernetes CLI - verify with `kubectl version --client` (for future functionality)

### Testing and Validation

- **Run tests**: `go test -v ./...` -- takes ~3 seconds. Set timeout to 60+ seconds for safety.
- **Run linter**: `~/go/bin/golangci-lint run` -- takes ~3 seconds. Set timeout to 60+ seconds for safety.
  - **Current State**: Linter reports 0 issues (clean codebase)
  - Always check that new code doesn't introduce linting violations

### Application Usage (Current State)

- **Get help**: `./ksail --help` (shows basic usage)
- **Get version**: `./ksail --version` (shows dev version info)
- **Current Status**: Only basic CLI structure implemented - no subcommands available yet
- **Planned Commands**: `init`, `up`, `down`, `list`, `reconcile` (not yet implemented)

## Validation Scenarios

### CRITICAL: Always test complete workflows after making changes

1. **Build Validation**:
   - Run `go build -o ksail .`
   - Verify `./ksail --help` shows usage
   - Verify `./ksail --version` shows version info

2. **Development Workflow** (Current Implementation):
   - Run tests: `go test -v ./...`
   - Check build: `go build -o ksail .`
   - Basic functionality: `./ksail --help`
   - Linting: `~/go/bin/golangci-lint run`

3. **Future Workflow Testing** (When Implemented):
   - Project initialization with `ksail init`
   - Cluster lifecycle with `ksail up/down`
   - Cluster listing with `ksail list`

## Build and Timing Information

### Command Timings (Measured on Current System)

- **`go build -o ksail .`**: <0.1 seconds when cached -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go test -v ./...`**: ~3 seconds -- SET TIMEOUT TO 60+ SECONDS for safety  
- **`~/go/bin/golangci-lint run`**: ~3 seconds -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go mod download`**: ~5 seconds -- SET TIMEOUT TO 60+ SECONDS for safety

### Linting Expectations

- **Current State**: 0 linting issues (clean codebase)
- **Config File**: `.golangci.yml` (note: .yml not .yaml)
- **Command**: `~/go/bin/golangci-lint run` (use full path, not system golangci-lint)
- **Focus**: Ensure new code doesn't introduce violations

## Codebase Navigation

### Current Directory Structure

- **`cmd/`**: CLI command implementations  
  - `root.go` - Main CLI setup with Cobra framework
  - `root_test.go` - Tests for root command
  - `ui/` - User interface utilities (notify, quiet, asciiart)
- **`main.go`**: Application entry point
- **`.golangci.yml`**: Linter configuration (v2 format)
- **`go.mod`**: Go module dependencies (cobra, fatih/color, go-snaps)

### Key Files and Patterns

- **CLI Framework**: Uses Cobra for command structure
- **Testing**: Uses go-snaps for snapshot testing
- **UI**: Colored output via fatih/color with symbols (✓, ✗, ⚠, ►)
- **Entry Point**: `main.go` creates root command and handles execution

### Planned Structure (Not Yet Implemented)

- **`cmd/init.go`**: Project scaffolding command
- **`cmd/up.go`**: Cluster provisioning command  
- **`cmd/down.go`**: Cluster destruction command
- **`cmd/list.go`**: Cluster listing command
- **`pkg/`**: Core business logic packages (future)
- **`internal/`**: Private application code (future)

## Configuration Files

### Current Configuration

- **`.golangci.yml`**: Comprehensive linting rules with depguard for import restrictions
- **`go.mod`**: Go 1.23.9+ with Cobra, color, and testing dependencies
- **`.github/workflows/`**: Minimal CI/CD setup (currently no-op CD)

### Build Configuration

- **No Makefile**: Uses standard Go commands
- **No Docker**: Pure Go build process  
- **Module**: `github.com/devantler-tech/ksail-go`

## Development Workflow

### Making Changes

1. **Always** validate current state first: `go test -v ./...`
2. **Always** build after changes: `go build -o ksail .`
3. **Always** test basic CLI: `./ksail --help`
4. **Always** run linter: `~/go/bin/golangci-lint run`
5. **Always** ensure tests pass before committing

### Adding New Features

1. **Follow existing patterns**: Study `cmd/root.go` and `cmd/ui/` structure
2. **Use Cobra framework**: For new CLI commands
3. **Add tests**: Follow snapshot testing pattern in `*_test.go` files
4. **Update help text**: Ensure CLI help stays accurate
5. **Test UI components**: Use existing notify/quiet/asciiart patterns

### Testing Strategies

- **Unit Tests**: Located in `*_test.go` files alongside source
- **Snapshot Testing**: Uses go-snaps for output validation
- **CLI Testing**: Test command execution and help output
- **Current Coverage**: Basic CLI structure and UI components only

## Troubleshooting

### Common Issues

- **Build fails**: Check Go version (need 1.24.6+), run `go mod download`
- **Linter fails**: Install with install script, use `~/go/bin/golangci-lint`
- **Import violations**: Check `.golangci.yml` depguard rules for allowed packages
- **Test failures**: Check snapshot files in `__snapshots__/` directories

### Known Limitations

- **Limited functionality**: Most CLI commands not yet implemented
- **No cluster operations**: Kubernetes functionality is planned, not implemented
- **Minimal testing**: Test coverage limited to existing components

## CI/CD Information

### GitHub Actions

- **CD**: Currently no-op (`.github/workflows/cd.yaml`)
- **CI**: Uses external workflow reference (likely in separate repo)
- **Future**: Will need comprehensive testing when functionality expands

### Development Approach

- **Work in Progress**: Core infrastructure exists, functionality being added incrementally
- **Clean Slate**: Current state has 0 linting issues, maintain this standard
- **Cobra Framework**: All CLI development should follow established Cobra patterns

---

**Last Updated**: Based on current repository state as of Go 1.24.6, minimal CLI implementation
