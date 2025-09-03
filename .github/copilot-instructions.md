# KSail Go - Kubernetes Cluster Management CLI

KSail is a Go-based CLI tool for managing local Kubernetes clusters and workloads declaratively. **Currently a work-in-progress migration from a previous implementation.** The core CLI structure exists but most functionality is planned/under development.

**This file configures the GitHub Copilot agent environment** to use the correct tools for linting (`mega-linter-runner -f go`), building (`go build`), and testing (`go test`) as specified in CONTRIBUTING.md.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build

- **Install Go 1.24.0+**: Verify with `go version` (project requires 1.24.0+ per go.mod, current runtime is 1.25.0)
- **Download dependencies**: `go mod download` (completes in ~0.03 seconds when cached, up to 30 seconds on first run)
- **Build the application**: `go build -o ksail .` -- takes ~0.77 seconds when dependencies cached. Set timeout to 60+ seconds for safety.
- **Install mega-linter-runner**: For comprehensive linting (primary linting tool): Install per [mega-linter docs](https://megalinter.io/latest/mega-linter-runner/#installation)
  - **CRITICAL**: Always use `mega-linter-runner -f go` for linting as specified in CONTRIBUTING.md
  - This is the primary linting tool used in CI and should be used locally for consistency

### Additional Development Tools (Optional)

- **mockery**: For generating mocks: Install per [mockery docs](https://vektra.github.io/mockery/v3.5/installation/)
  - Configuration in `.mockery.yml` - supports `mockery` command to regenerate mocks
  - **Current Environment**: Available at `/home/runner/go/bin/mockery` - takes ~0.78 seconds to run
- **golangci-lint**: Alternative linting tool: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/go/bin latest`
  - **NOTE**: Project primarily uses mega-linter, but golangci-lint may be used as fallback
  - **Current Environment**: Available at `~/go/bin/golangci-lint` - takes ~34 seconds to run
  - Use `~/go/bin/golangci-lint run` if using this tool

### Required Dependencies

- **Go 1.24.0+**: Programming language runtime (required)
- **Docker**: Container runtime - verify with `docker --version` (for future functionality)
  - **Current Environment**: Docker 28.0.4 available
- **Kind**: Local Kubernetes clusters - verify with `kind version` (for future functionality)
  - **Current Environment**: Kind v0.30.0 available
- **kubectl**: Kubernetes CLI - verify with `kubectl version --client` (for future functionality)
  - **Current Environment**: kubectl v1.34.0 available

### Testing and Validation

- **Run tests**: `go test -v ./...` -- takes ~1m11s. Set timeout to 120+ seconds for safety.
- **Generate mocks**: `mockery` -- regenerates test mocks in ~0.78 seconds. Required when interface definitions change.
- **Run linter**: `mega-linter-runner -f go` -- comprehensive linting with Go flavor as specified in CONTRIBUTING.md
  - **Primary linting tool**: This is the main linting tool used in CI and locally
  - Configuration in `.mega-linter.yml` with `APPLY_FIXES: all`
  - **Auto-fix capability**: Automatically fixes formatting and style issues when run
  - **IMPORTANT**: Takes 2-3 minutes to complete due to comprehensive security scanning and link checking
  - **Network dependencies**: May fail with timeouts if external links are unreachable (non-critical for code quality)
  - **Alternative**: `~/go/bin/golangci-lint run` if mega-linter-runner not available (takes ~34 seconds)
  - **Current State**: Clean codebase with 0 linting issues (mega-linter auto-fixes enabled)
  - Always check that new code doesn't introduce linting violations

### Application Usage (Current State)

- **Get help**: `./ksail --help` (shows basic usage)
- **Get version**: `./ksail --version` (shows dev version info)
- **Current Status**: CLI structure implemented with working stub commands
- **Available Commands**: `init`, `up`, `down`, `start`, `stop`, `list`, `status`, `reconcile` (all implemented as working stubs)

## Validation Scenarios

### CRITICAL: Always test complete workflows after making changes

1. **Build Validation**: Run `go build -o ksail .`, verify `./ksail --help` shows usage, verify `./ksail --version` shows version info

2. **Development Workflow**: Run tests: `go test -v ./...`, check build: `go build -o ksail .`, basic functionality: `./ksail --help`, linting: `mega-linter-runner -f go`

3. **Command Testing** (All commands work as stubs): Project initialization with `ksail init`, cluster lifecycle with `ksail up/down`, cluster listing with `ksail list`

4. **MANDATORY Full Workflow Test**: After any changes, run this complete validation:

   ```bash
   # Complete development validation workflow
   go test -v ./...                    # ~1m11s - ALL tests must pass
   go build -o ksail .                 # ~0.77 seconds - must build successfully
   ./ksail --help                      # Must show help without errors
   ./ksail --version                   # Must show version info
   mega-linter-runner -f go            # Primary linting tool - takes 2-3 minutes, auto-fixes issues
   # Alternative if mega-linter not available: ~/go/bin/golangci-lint run (takes ~34 seconds)

   # Test core functionality
   ./ksail init --container-engine Docker --distribution Kind
   ./ksail up && ./ksail status && ./ksail list && ./ksail down
   ```

5. **Additional CLI Testing**: Test command help and error handling:

   ```bash
   # Test all command help outputs
   ./ksail init --help
   ./ksail up --help
   ./ksail down --help
   ./ksail status --help
   ./ksail list --help
   ./ksail start --help
   ./ksail stop --help
   ./ksail reconcile --help
   
   # Test different flag combinations
   ./ksail init --container-engine Podman --distribution K3d
   ./ksail list --all
   ```

## Build and Timing Information

### Command Timings (Measured on Current System)

- **`go build -o ksail .`**: ~0.77 seconds when cached, ~1s first time -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go test -v ./...`**: ~1m11s -- SET TIMEOUT TO 120+ SECONDS for safety
- **`mega-linter-runner -f go`**: Primary linting tool, takes 2-3 minutes (runs multiple linters and security scanners) -- SET TIMEOUT TO 300+ SECONDS for safety
- **`~/go/bin/golangci-lint run`**: ~34 seconds (alternative linter) -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go mod download`**: ~0.03 seconds when cached, up to 30 seconds first run -- SET TIMEOUT TO 60+ SECONDS for safety
- **`mockery`**: ~0.78 seconds for mock generation -- SET TIMEOUT TO 60+ SECONDS for safety
- **NEVER CANCEL**: All commands may take longer on different systems. Always wait for completion.

### Linting Expectations

- **Current State**: Clean codebase with 0 linting issues (mega-linter auto-fixes enabled)
- **Primary Tool**: `mega-linter-runner -f go` with configuration in `.mega-linter.yml`
- **Alternative Tool**: `~/go/bin/golangci-lint run` with config in `.golangci.yml` (note: .yml extension, not .yaml)
- **CONTRIBUTING.md Requirement**: Must use `mega-linter-runner -f go` for consistency with CI
- **Focus**: Ensure new code doesn't introduce additional violations
- **Network Issues**: Mega-linter includes link checking which may timeout on external URLs (not a code quality issue)
- **Performance**: For faster iteration during development, use `~/go/bin/golangci-lint run` for Go-specific linting (~34 seconds)

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
- **Repository Size**: ~20,173 lines of Go code across 90 files

### Current Implementation Status

- **CLI Commands**: All major commands implemented as working stubs (init, up, down, start, stop, list, status, reconcile)
- **Package Structure**: Proper separation with cmd/, pkg/, internal/ directories
- **Cluster Provisioning**: Real implementation in pkg/provisioner/cluster with Kind integration
- **Testing**: Comprehensive test coverage (38 test files) with mocks and snapshot testing
- **Future Development**: Stubs provide framework for full implementation

## Configuration Files

### Current Configuration

- **`.mega-linter.yml`**: Primary linting configuration with `APPLY_FIXES: all`
- **`.golangci.yml`**: Alternative comprehensive linting rules with depguard for import restrictions
- **`go.mod`**: Go 1.24.0+ with Cobra, color, testing, Docker, and Kind dependencies
- **`.github/workflows/`**: Complex CI/CD with matrix testing across container engines and distributions
- **`.mockery.yml`**: Mockery configuration for generating test mocks

### Build Configuration

- **No Makefile**: Uses standard Go commands
- **No Docker**: Pure Go build process
- **Module**: `github.com/devantler-tech/ksail-go`

## Development Workflow

### Making Changes

1. **Always** validate current state first: `go test -v ./...`
2. **Always** build after changes: `go build -o ksail .`
3. **Always** test basic CLI: `./ksail --help`
4. **Always** run linter: `mega-linter-runner -f go` (primary) or `~/go/bin/golangci-lint run` (fallback)
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

- **Build fails**: Check Go version (need 1.24.0+), run `go mod download`
- **Linter fails**: Install mega-linter-runner per [docs](https://megalinter.io/latest/mega-linter-runner/#installation), or use golangci-lint as fallback
- **Import violations**: Check `.mega-linter.yml` and `.golangci.yml` configuration for allowed packages
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

**Last Updated**: Based on current repository state as of Go 1.25.0, CLI stub implementation. Validated on GitHub Actions environment with Docker 28.0.4, Kind v0.30.0, kubectl v1.34.0. All timings measured and commands tested to ensure accuracy of instructions.
