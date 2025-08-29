# KSail Go - Kubernetes Cluster Management CLI

KSail is a Go-based CLI tool for managing local Kubernetes clusters and workloads declaratively. **Currently a work-in-progress migration from a previous implementation.** The core CLI structure exists but most functionality is planned/under development.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build

- **Install Go 1.23.9+**: Verify with `go version` (project requires 1.23.9+ per go.mod, runtime is 1.24.6)
- **Download dependencies**: `go mod download` (completes in ~15 seconds)
- **Build the application**: `go build -o ksail .` -- takes ~11 seconds when dependencies cached. Set timeout to 60+ seconds for safety.
- **Install golangci-lint v2**: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/go/bin latest`
  - **CRITICAL**: Always use `~/go/bin/golangci-lint run`, not system golangci-lint

### Additional Development Tools (Optional)

- **mega-linter**: For comprehensive linting (used in CONTRIBUTING.md): Install per [mega-linter docs](https://megalinter.io/latest/mega-linter-runner/#installation)
- **mockery**: For generating mocks: Install per [mockery docs](https://vektra.github.io/mockery/v3.5/installation/)
  - Configuration in `.mockery.yml` - supports `mockery` command to regenerate mocks
  - **Current Environment**: Not installed by default - install only if modifying interface code

### Required Dependencies

- **Go 1.23.9+**: Programming language runtime (required)
- **Docker**: Container runtime - verify with `docker --version` (for future functionality)
  - **Current Environment**: Docker 28.0.4 available
- **Kind**: Local Kubernetes clusters - verify with `kind version` (for future functionality)
  - **Current Environment**: Kind v0.29.0 available
- **kubectl**: Kubernetes CLI - verify with `kubectl version --client` (for future functionality)
  - **Current Environment**: kubectl v1.33.4 available

### Testing and Validation

- **Run tests**: `go test -v ./...` -- takes ~33 seconds. Set timeout to 60+ seconds for safety.
- **Run linter**: `~/go/bin/golangci-lint run` -- takes ~14 seconds. Set timeout to 60+ seconds for safety.
  - **Current State**: Linter reports 0 issues (clean codebase)
  - Always check that new code doesn't introduce linting violations

### Application Usage (Current State)

- **Get help**: `./ksail --help` (shows basic usage)
- **Get version**: `./ksail --version` (shows dev version info)
- **Current Status**: CLI structure implemented with working stub commands
- **Available Commands**: `init`, `up`, `down`, `start`, `stop`, `list`, `status`, `reconcile` (all implemented as working stubs)

## Validation Scenarios

### CRITICAL: Always test complete workflows after making changes

1. **Build Validation**: Run `go build -o ksail .`, verify `./ksail --help` shows usage, verify `./ksail --version` shows version info

2. **Development Workflow**: Run tests: `go test -v ./...`, check build: `go build -o ksail .`, basic functionality: `./ksail --help`, linting: `~/go/bin/golangci-lint run`

3. **Command Testing** (All commands work as stubs): Project initialization with `ksail init`, cluster lifecycle with `ksail up/down`, cluster listing with `ksail list`

4. **MANDATORY Full Workflow Test**: After any changes, run this complete validation:

   ```bash
   # Complete development validation workflow
   go test -v ./...              # ~33 seconds - ALL tests must pass
   go build -o ksail .           # ~11 seconds - must build successfully
   ./ksail --help               # Must show help without errors
   ./ksail --version            # Must show version info
   ~/go/bin/golangci-lint run   # ~15 seconds - must show 0 issues (ignore wsl warnings)

   # Test core functionality
   ./ksail init --container-engine Docker --distribution Kind
   ./ksail up && ./ksail status && ./ksail list && ./ksail down
   ```

## Build and Timing Information

### Command Timings (Measured on Current System)

- **`go build -o ksail .`**: ~11 seconds when cached, ~12s first time -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go test -v ./...`**: ~33 seconds -- SET TIMEOUT TO 60+ SECONDS for safety
- **`~/go/bin/golangci-lint run`**: ~15 seconds -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go mod download`**: ~15 seconds -- SET TIMEOUT TO 60+ SECONDS for safety
- **NEVER CANCEL**: All commands may take longer on different systems. Always wait for completion.

### Linting Expectations

- **Current State**: 0 linting issues (clean codebase)
- **Config File**: `.golangci.yml` (note: .yml extension, not .yaml)
- **Command**: `~/go/bin/golangci-lint run` (use full path, not system golangci-lint)
- **Focus**: Ensure new code doesn't introduce violations

## Codebase Navigation

### Current Directory Structure

- **`cmd/`**: CLI command implementations
  - `root.go` - Main CLI setup with Cobra framework
  - `root_test.go` - Tests for root command
  - `init.go`, `up.go`, `down.go`, `list.go`, etc. - Working stub command implementations
  - `ui/` - User interface utilities (notify, quiet, asciiart)
- **`pkg/`**: Core business logic packages
  - `provisioner/cluster/` - Kubernetes cluster provisioning logic (real implementation with mocks)
- **`internal/`**: Private application code
  - `utils/path/` - Path utility functions
- **`main.go`**: Application entry point
- **`.golangci.yml`**: Linter configuration (v2 format)
- **`go.mod`**: Go module dependencies (cobra, fatih/color, go-snaps, docker, kind)

### Key Files and Patterns

- **CLI Framework**: Uses Cobra for command structure
- **Testing**: Uses go-snaps for snapshot testing
- **UI**: Colored output via fatih/color with symbols (✓, ✗, ⚠, ►)
- **Entry Point**: `main.go` creates root command and handles execution
- **Repository Size**: ~16,300 lines of Go code across 62 files

### Current Implementation Status

- **CLI Commands**: All major commands implemented as working stubs (init, up, down, start, stop, list, status, reconcile)
- **Package Structure**: Proper separation with cmd/, pkg/, internal/ directories
- **Cluster Provisioning**: Real implementation in pkg/provisioner/cluster with Kind integration
- **Testing**: Comprehensive test coverage (26 tests) with mocks and snapshot testing
- **Future Development**: Stubs provide framework for full implementation

## Configuration Files

### Current Configuration

- **`.golangci.yml`**: Comprehensive linting rules with depguard for import restrictions
- **`go.mod`**: Go 1.23.9+ with Cobra, color, testing, Docker, and Kind dependencies
- **`.github/workflows/`**: Complex CI/CD with matrix testing across container engines and distributions
- **`.mockery.yml`**: Mockery configuration for generating test mocks
- **`.mega-linter.yml`**: Mega-linter configuration (alternative to golangci-lint)

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
- **Mock Testing**: pkg/provisioner/cluster uses mocks for Docker API testing
- **Current Coverage**: CLI structure, UI components, and cluster provisioning logic

## Troubleshooting

### Common Issues

- **Build fails**: Check Go version (need 1.23.9+), run `go mod download`
- **Linter fails**: Install with install script, use `~/go/bin/golangci-lint`
- **Linter warnings about deprecated 'wsl'**: Expected deprecation warnings about wsl linter, can be ignored. Linter suggests migrating to wsl_v5 but current configuration works correctly.
- **Import violations**: Check `.golangci.yml` depguard rules for allowed packages
- **Test failures**: Check snapshot files in `__snapshots__/` directories
- **Mock generation**: Run `mockery` to regenerate mocks if tests fail due to interface changes

### Known Limitations

- **Stub implementations**: Most CLI commands return success messages but don't perform real operations
- **Work in progress**: Core infrastructure exists, full functionality being added incrementally
- **System dependencies**: Real cluster operations will require Docker, Kind, kubectl

## CI/CD Information

### GitHub Actions

- **CI**: Uses external reusable workflow (`devantler-tech/reusable-workflows`) for Go CI/CD
- **System Tests**: Matrix testing across container engines (Docker/Podman) and distributions (Kind/K3d)
  - Tests all CLI commands: `init`, `up`, `status`, `list`, `start`, `reconcile`, `down`
  - Runs with different configurations: `--container-engine Docker/Podman --distribution Kind/K3d`
  - Cannot be run locally, only in CI environment
- **Coverage**: Codecov integration for test coverage reporting
- **Validation**: Tests all CLI commands in realistic scenarios

### Development Approach

- **Work in Progress**: Core infrastructure exists, functionality being added incrementally
- **Clean Slate**: Current state has 0 linting issues, maintain this standard
- **Cobra Framework**: All CLI development should follow established Cobra patterns

---

**Last Updated**: Based on current repository state as of Go 1.24.6, CLI stub implementation. Validated on GitHub Actions environment with Docker 28.0.4, Kind v0.29.0, kubectl v1.33.4.
