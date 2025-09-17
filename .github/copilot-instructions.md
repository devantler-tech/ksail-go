# KSail Go - Kubernetes Cluster Management CLI

KSail is a Go-based CLI tool for managing local Kubernetes clusters and workloads declaratively. **Currently a work-in-progress migration from a previous implementation.** The core CLI structure exists but most functionality is planned/under development.

**This file configures the GitHub Copilot agent environment** to use the correct tools for linting (`mega-linter-runner -f go`), building (`go build`), and testing (`go test`) as specified in CONTRIBUTING.md.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## GitHub Copilot Agent Configuration

This repository follows [Best practices for Copilot coding agent in your repository](https://gh.io/copilot-coding-agent-tips) with the following specific considerations:

### Context Files and Documentation

- **Primary Instructions**: This `.github/copilot-instructions.md` file provides comprehensive repository context
- **Package Documentation**: README.md files throughout the repository provide detailed package-specific documentation and serve as the primary source for understanding individual components
- **Allowed Modifications**: Copilot agents may add or rewrite context files according to best practices if specific code areas need more explicit instructions
- **Context Principle**: Provide explicit context rather than relying on inference to ensure consistent behavior across different agents

## Working Effectively

### Bootstrap and Build

- **Install Go 1.24.0+**: Verify with `go version` (project requires 1.24.0+ per go.mod, current runtime is 1.25.1)
- **Download dependencies**: `go mod download` (completes in ~0.03 seconds when cached, up to 30 seconds on first run)
- **Build the application**: `go build -o ksail .` -- takes ~0.5 seconds first time, ~0.1 seconds when cached. Set timeout to 60+ seconds for safety.
- **Install mega-linter-runner**: For comprehensive linting (primary linting tool): Install per [mega-linter docs](https://megalinter.io/latest/mega-linter-runner/#installation)
  - **CRITICAL**: Always use `mega-linter-runner -f go` for linting as specified in CONTRIBUTING.md
  - This is the primary linting tool used in CI and should be used locally for consistency

### Additional Development Tools (Optional)

- **mockery**: For generating mocks: Install per [mockery docs](https://vektra.github.io/mockery/v3.5/installation/)
  - Configuration in `.mockery.yml` - supports `mockery` command to regenerate mocks
  - **Current Environment**: Available at `/home/runner/go/bin/mockery` - takes ~1.3 seconds to run
- **golangci-lint**: Alternative linting tool: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/go/bin latest`
  - **NOTE**: Project primarily uses mega-linter, but golangci-lint may be used as fallback
  - **Current Environment**: Available at `~/go/bin/golangci-lint` - takes ~1m16s to run when clean
  - **Performance Notes**:
    - Takes longer when issues are present (proportional to number of violations)
    - Use `~/go/bin/golangci-lint run --fast-only` for quick feedback on fast linters only
    - Full run recommended before commits: `~/go/bin/golangci-lint run`

### Required Dependencies

- **Go 1.24.0+**: Programming language runtime (required)
- **Docker**: Container runtime - verify with `docker --version` (for future functionality)
  - **Current Environment**: Docker 28.0.4 available
- **Kind**: Local Kubernetes clusters - verify with `kind version` (for future functionality)
  - **Current Environment**: Kind v0.30.0 available
- **kubectl**: Kubernetes CLI - verify with `kubectl version --client` (for future functionality)
  - **Current Environment**: kubectl v1.34.0 available

### Testing and Validation

- **Run tests**: `go test -v ./...` -- takes ~37 seconds. Set timeout to 120+ seconds for safety.
- **Generate mocks**: `mockery` -- regenerates test mocks in ~1.3 seconds. Required when interface definitions change.
- **Run linter**: `mega-linter-runner -f go` -- comprehensive linting with Go flavor as specified in CONTRIBUTING.md
  - **Primary linting tool**: This is the main linting tool used in CI and locally
  - Configuration in `.mega-linter.yml` with `APPLY_FIXES: all` and `DISABLE_LINTERS: - GO_GOLANGCI_LINT`
  - **Auto-fix capability**: Automatically fixes formatting and style issues when run
  - **IMPORTANT**: Takes 11-12 minutes to complete due to comprehensive security scanning and link checking (set timeout to 1200+ seconds)
  - **Network dependencies**: May fail with timeouts if external links are unreachable (non-critical for code quality)
  - **Go-specific linting**: Run `~/go/bin/golangci-lint run` separately for Go-specific checks (takes ~1m16s, shows 0 issues)
  - **Current CI State**: Mega-linter passes (golangci-lint disabled), separate golangci-lint run shows 0 issues
  - Always check that new code doesn't introduce additional linting violations in enabled linters

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
   go test -v ./...                    # ~37 seconds - ALL tests must pass
   go build -o ksail .                 # ~0.5s first time, ~0.1s cached - must build successfully
   ./ksail --help                      # Must show help without errors
   ./ksail --version                   # Must show version info
   mega-linter-runner -f go            # Primary linting tool - takes 11-12 minutes, auto-fixes issues, CI-consistent
   # Optional Go-specific linting: ~/go/bin/golangci-lint run (takes ~1m16s, shows 0 issues)

   # Test core functionality
   ./ksail init --distribution Kind
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
   ./ksail init --distribution K3d
   ./ksail list --all
   ```

## Build and Timing Information

### Command Timings (Measured on Current System)

- **`go build -o ksail .`**: ~0.5 seconds first time, ~0.1 seconds when cached -- SET TIMEOUT TO 60+ SECONDS for safety
- **`go test -v ./...`**: ~37 seconds -- SET TIMEOUT TO 120+ SECONDS for safety
- **`mega-linter-runner -f go`**: Primary linting tool, takes 11-12 minutes (runs multiple linters and security scanners) -- SET TIMEOUT TO 1200+ SECONDS for safety
- **`~/go/bin/golangci-lint run`**: ~1m16s (alternative linter) -- SET TIMEOUT TO 120+ SECONDS for safety
- **`go mod download`**: ~0.03 seconds when cached, up to 30 seconds first run -- SET TIMEOUT TO 60+ SECONDS for safety
- **`mockery`**: ~1.3 seconds for mock generation -- SET TIMEOUT TO 60+ SECONDS for safety
- **NEVER CANCEL**: All commands may take longer on different systems. Always wait for completion.

### Linting Expectations

- **CI Status**: CI passes because golangci-lint is disabled in mega-linter configuration (`.mega-linter.yml` has `DISABLE_LINTERS: - GO_GOLANGCI_LINT`)
- **Direct golangci-lint**: When run separately, `~/go/bin/golangci-lint run` shows 0 issues
- **Primary Tool**: `mega-linter-runner -f go` with configuration in `.mega-linter.yml` (excludes golangci-lint)
- **Go-specific Linting**: Use `~/go/bin/golangci-lint run` separately for Go-specific linting with config in `.golangci.yml`
- **CONTRIBUTING.md Requirement**: Must use `mega-linter-runner -f go` for consistency with CI
- **Current State**: Clean codebase with both mega-linter and golangci-lint passing successfully
- **Focus**: Ensure new code doesn't introduce additional violations in enabled linters
- **Network Issues**: Mega-linter includes link checking which may timeout on external URLs (not a code quality issue)
- **Performance**: For faster iteration during development, use `~/go/bin/golangci-lint run` for Go-specific linting (~1m16s)
- **Quick feedback**: Use `~/go/bin/golangci-lint run --fast-only` for rapid iteration on fast linters only

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
- **Repository Size**: ~22,792 lines of Go code across 90 files

### Current Implementation Status

- **CLI Commands**: All major commands implemented as working stubs (init, up, down, start, stop, list, status, reconcile)
- **Package Structure**: Proper separation with cmd/, pkg/, internal/ directories
- **Cluster Provisioning**: Real implementation in pkg/provisioner/cluster with Kind integration
- **Testing**: Comprehensive test coverage (39 test files) with mocks and snapshot testing
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
4. **Always** run linter: `mega-linter-runner -f go` (primary, CI-consistent) and optionally `~/go/bin/golangci-lint run` (Go-specific)
5. **Always** ensure tests pass before committing
6. **Pre-commit hooks**: Be aware that pre-commit hooks are installed and will run automatically
   - Hooks include golangci-lint formatting and mockery generation
   - **Important**: Commits may take longer due to hook execution - wait for completion
   - **Cannot assume immediate success**: Hooks may fail and prevent commit - check for errors
   - See `.pre-commit-config.yaml` for configured hooks

### Adding New Features

1. **Follow existing patterns**: Study `cmd/root.go` and `cmd/ui/` structure
2. **Use Cobra framework**: For new CLI commands
3. **Add tests**: Follow snapshot testing pattern in `*_test.go` files
4. **Update help text**: Ensure CLI help stays accurate
5. **Test UI components**: Use existing notify/quiet/asciiart patterns

### Testing Strategies

- **Unit Tests**: Located in `*_test.go` files alongside source
- **Snapshot Testing**: **CRITICAL FOR CLI COMMANDS** - Uses go-snaps for output validation in cmd/ directory
  - All CLI command tests in `cmd/*_test.go` use `snaps.MatchSnapshot(t, output)` for consistent output validation
  - Snapshot files stored in `cmd/__snapshots__/` directory (e.g., `root_test.snap`, `init_test.snap`)
  - **TestMain function required**: Each cmd test file needs `snaps.Clean(m, snaps.CleanOpts{Sort: true})` in TestMain
  - **Regenerate snapshots**: Run tests with `UPDATE_SNAPSHOTS=true go test ./cmd/...` to update expected output
  - **Essential for CLI changes**: Any changes to command output, help text, or error messages require snapshot updates
- **CLI Testing**: Test command execution and help output
- **Mock Testing**: pkg/provisioner/cluster uses mocks for Docker API testing
- **Current Coverage**: CLI structure, UI components, and cluster provisioning logic

### Test Naming Conventions and Structure

**CRITICAL NAMING REQUIREMENTS**: Follow strict Go community naming conventions for all test functions:

#### Test Function Naming Rules

- **Primary Pattern**: `func TestXxx(t *testing.T)` where `Xxx` is the **method/function/constructor name only**
- **NEVER include struct names**: `TestManagerLoadConfig` ❌ → `TestLoadConfig` ✅
- **NEVER include explanations**: `TestLoadConfigWithValidInput` ❌ → `TestLoadConfig` ✅ (use subtests instead)
- **Method names only**: For `type Manager struct` with `func (m *Manager) LoadConfig()`, test should be `TestLoadConfig`
- **Constructor pattern**: For `func NewManager()`, test should be `TestNewManager`

#### One Test Per Method/Function Rule

- **Preferred**: Single test function per method/function/constructor using table-driven tests or `t.Run` subtests
- **Example structure**:

  ```go
  func TestLoadConfig(t *testing.T) {
      testCases := []struct {
          name string
          // test case fields
      }{
          {name: "valid config", /* ... */},
          {name: "invalid config", /* ... */},
          {name: "missing config", /* ... */},
      }
      
      for _, tc := range testCases {
          t.Run(tc.name, func(t *testing.T) {
              // test implementation
          })
      }
  }
  ```

#### Test Organization Principles

- **One test per method**: Strive for single test function per method under test
- **Use t.Run for scenarios**: Multiple test scenarios should use `t.Run` subtests within single test function
- **Table-driven preferred**: Use table-driven tests for multiple input/output combinations
- **No struct prefixes**: Never include struct/type names in test function names
- **Focus on behavior**: Test function names should reflect the method being tested, not the implementation details

### Coverage Expectations

- **testutils packages**: Should NOT have code coverage as they are test utility packages
  - `cmd/internal/testutils/`: Test utilities for command testing (no coverage expected)
  - `internal/testutils/`: Shared test utilities (no coverage expected)
  - `pkg/*/testutils/`: Package-specific test utilities (no coverage expected)
- **Production code**: Should have high test coverage with proper unit tests
- **Focus coverage efforts**: On actual implementation code, not test helper utilities

## Troubleshooting

### Common Issues

- **Build fails**: Check Go version (need 1.24.0+), run `go mod download`
- **Linter fails**: Install mega-linter-runner per [docs](https://megalinter.io/latest/mega-linter-runner/#installation). Note: golangci-lint is disabled in mega-linter config but can be run separately
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
- **System Tests**: Matrix testing across different configurations
  - Tests all CLI commands: `init`, `up`, `status`, `list`, `start`, `reconcile`, `down`
  - Runs with different configurations: `--distribution Kind/K3d/EKS`
  - Cannot be run locally, only in CI environment
- **Coverage**: Codecov integration for test coverage reporting
- **Validation**: Tests all CLI commands in realistic scenarios

### Development Approach

- **Work in Progress**: Core infrastructure exists, functionality being added incrementally
- **Clean Slate**: Current state has CI-passing configuration with mega-linter, separate golangci-lint shows 0 issues
- **Cobra Framework**: All CLI development should follow established Cobra patterns

---

**Last Updated**: Based on current repository state as of Go 1.25.1, CLI stub implementation. Validated on GitHub Actions environment with Docker 28.0.4, Kind v0.30.0, kubectl v1.34.0. All timings measured and commands tested to ensure accuracy of instructions. Updated with comprehensive validation of all commands and current golangci-lint status (0 issues). Build times updated based on real measurements: ~0.5s first time, ~0.1s when cached.
