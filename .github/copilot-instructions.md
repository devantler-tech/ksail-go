# KSail Go - Kubernetes Cluster Management CLI

KSail is a Go-based CLI tool for managing local Kubernetes clusters and workloads declaratively. **Currently a work-in-progress migration from a previous implementation.** The core CLI structure exists but most functionality is planned/under development.

**This file configures the GitHub Copilot agent environment** to use the correct tools for linting (`mega-linter-runner -f go`), building (`go build`), and testing (`go test`) as specified in CONTRIBUTING.md.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Quick Start for New Agents

**CRITICAL**: Start here for any work in this repository:

1. **Navigate to repository root**: `cd /path/to/ksail-go` (ensure you're in the directory containing `go.mod`)
2. **Verify environment**: `go version` (must be 1.24.0+, current: 1.25.1)
3. **Download dependencies**: `go mod download` (~0.03s if cached, up to 30s first run)
4. **Build application**: `go build -o ksail .` (~0.5s first time, ~0.1s cached)
5. **Verify build**: `./ksail --help` (should show CLI help without errors)
6. **Run tests**: `go test -v ./...` (~37s, must pass before making changes)

If any step fails, see the [Troubleshooting](#troubleshooting) section before proceeding.

## GitHub Copilot Agent Configuration

This repository follows [Best practices for Copilot coding agent in your repository](https://gh.io/copilot-coding-agent-tips) with the following specific considerations:

### Context Files and Documentation

- **Primary Instructions**: This `.github/copilot-instructions.md` file provides comprehensive repository context
- **Package Documentation**: README.md files throughout the repository provide detailed package-specific documentation and serve as the primary source for understanding individual components
- **Allowed Modifications**: Copilot agents may add or rewrite context files according to best practices if specific code areas need more explicit instructions
- **Context Principle**: Provide explicit context rather than relying on inference to ensure consistent behavior across different agents

## Working Effectively

**IMPORTANT**: All commands below assume you are in the repository root directory (where `go.mod` is located).

### Essential Tools (Install First)

These tools are required for basic development workflow:

1. **Go 1.24.0+**: Verify with `go version` (project requires 1.24.0+ per go.mod, current runtime is 1.25.1)
2. **mega-linter-runner**: For primary linting (REQUIRED for CI consistency)

   ```bash
   # Install per: https://megalinter.io/latest/mega-linter-runner/#installation
   npm install mega-linter-runner --global
   # OR use npx: npx mega-linter-runner -f go
   ```

3. **golangci-lint**: Go-specific linting and fast feedback

   ```bash
   # Install: 
   curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/go/bin latest
   ```

   - **Current Environment**: Available at `~/go/bin/golangci-lint` - takes ~1m16s to run when clean
   - **Usage**: `~/go/bin/golangci-lint run` for full check, `~/go/bin/golangci-lint run --fast-only` for quick feedback

### Bootstrap and Build

**Step-by-step build process:**

1. **Download dependencies**: `go mod download` (completes in ~0.03 seconds when cached, up to 30 seconds on first run)
2. **Build the application**: `go build -o ksail .` -- takes ~0.5 seconds first time, ~0.1 seconds when cached. Set timeout to 60+ seconds for safety.
3. **Verify build success**: `./ksail --help` should display CLI help without errors

**If build fails**: Check Go version, run `go mod tidy`, then retry build.

### Optional Development Tools

These tools enhance the development experience but are not required for basic functionality:

- **mockery**: For generating test mocks

  ```bash
  # Install per: https://vektra.github.io/mockery/v3.5/installation/
  go install github.com/vektra/mockery/v3@latest
  ```

  - Configuration in `.mockery.yml` - run `mockery` to regenerate mocks (~1.3s)
  - **Current Environment**: Available at `/home/runner/go/bin/mockery`
  - **When to use**: Required when interface definitions change

### System Dependencies (For Future Functionality)

These are available in the current environment but not required for current stub implementations:

- **Docker 28.0.4**: Container runtime - verify with `docker --version`
- **Kind v0.30.0**: Local Kubernetes clusters - verify with `kind version`  
- **kubectl v1.34.0**: Kubernetes CLI - verify with `kubectl version --client`

### Core Development Commands

**Test execution** (ALWAYS run before making changes):

```bash
go test -v ./...                    # ~37 seconds - ALL tests must pass
```

**Mock generation** (when interfaces change):

```bash
mockery                             # ~1.3 seconds - regenerates all mocks
```

**Primary linting** (REQUIRED for CI consistency):

```bash
mega-linter-runner -f go            # ~11-12 minutes - comprehensive scanning
```

- **CRITICAL**: This is the main linting tool specified in CONTRIBUTING.md
- Configuration in `.mega-linter.yml` with `APPLY_FIXES: all`
- **Auto-fixes** formatting and style issues
- **Network dependencies**: May timeout on external links (non-critical for code quality)
- **Expected timeouts**: Link checking and security scanning take time - this is normal

**Fast Go-specific linting** (optional for quick feedback):

```bash
~/go/bin/golangci-lint run          # ~1m16s - Go-specific checks only
~/go/bin/golangci-lint run --fast-only  # Quick feedback on fast linters
```

### Application Usage (Current State)

**Basic CLI testing:**

```bash
./ksail --help                      # Shows basic usage and available commands
./ksail --version                   # Shows dev version info
```

**Available Commands** (all implemented as working stubs):

- `init` - Initialize a new project with distribution selection
- `up` - Start the Kubernetes cluster (stub)
- `down` - Destroy a cluster (stub)  
- `start` - Start a stopped cluster (stub)
- `stop` - Stop the Kubernetes cluster (stub)
- `list` - List clusters (stub)
- `status` - Show status of the Kubernetes cluster (stub)
- `reconcile` - Reconcile workloads in the cluster (stub)

**Example workflow:**

```bash
./ksail init --distribution Kind    # Initialize project
./ksail up                          # Start cluster (stub)
./ksail status                      # Check status (stub)
./ksail list                        # List clusters (stub)
./ksail down                        # Destroy cluster (stub)
```

## Validation Scenarios

### CRITICAL: Always test complete workflows after making changes

**Before starting any development work:**

1. Ensure you're in repository root (contains `go.mod`)
2. Run `go test -v ./...` to verify current state
3. Run `go build -o ksail .` to ensure clean build
4. Run `./ksail --help` to verify CLI functionality

### Standard Development Workflow

**After making ANY code changes, run this validation sequence:**

```bash
# 1. Test first (MANDATORY - must pass)
go test -v ./...                    # ~37 seconds - ALL tests must pass

# 2. Build verification
go build -o ksail .                 # ~0.5s first time, ~0.1s cached

# 3. Basic functionality check
./ksail --help                      # Must show help without errors
./ksail --version                   # Must show version info

# 4. Lint (choose one based on time available)
mega-linter-runner -f go            # Full lint: ~11-12 minutes (CI-consistent)
# OR for faster feedback:
~/go/bin/golangci-lint run          # Go-specific: ~1m16s

# 5. CLI functionality test
./ksail init --distribution Kind
./ksail up && ./ksail status && ./ksail list && ./ksail down
```

**If any step fails:**

- **Tests fail**: Fix failing tests before proceeding
- **Build fails**: Check Go version, run `go mod tidy`, retry
- **CLI fails**: Check for compilation errors, missing dependencies
- **Lint fails**: Review linter output, fix violations or justify exceptions

### Comprehensive Testing Scenarios

**1. Build Validation**

```bash
go build -o ksail .                 # Must complete without errors
./ksail --help                      # Must show usage
./ksail --version                   # Must show version info
```

**2. Command Testing** (All commands work as stubs)

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
./ksail init --distribution EKS
./ksail list --all
```

**3. Mock Generation Testing** (when interfaces change)

```bash
mockery                             # Must complete without errors (~1.3s)
go test -v ./...                    # Verify mocks work with tests
```

## Command Reference and Timing

### Essential Command Timings (Measured and Validated)

**CRITICAL**: Always set appropriate timeouts to prevent premature cancellation.

| Command                      | Time                               | Timeout | Notes                     |
|------------------------------|------------------------------------|---------|---------------------------|
| `go mod download`            | ~0.03s cached, up to 30s first run | 60s     | Dependency download       |
| `go build -o ksail .`        | ~0.5s first time, ~0.1s cached     | 60s     | Build application         |
| `go test -v ./...`           | ~37s                               | 120s    | Full test suite           |
| `mockery`                    | ~1.3s                              | 60s     | Mock generation           |
| `~/go/bin/golangci-lint run` | ~1m16s                             | 120s    | Go-specific linting       |
| `mega-linter-runner -f go`   | ~11-12 minutes                     | 1200s   | Full lint + security scan |

**NEVER CANCEL**: All commands may take longer on different systems. Always wait for completion.

### Working Directory Requirements

**ALL commands assume you are in the repository root** (directory containing `go.mod`).

If commands fail with "go.mod not found" or similar:

```bash
cd /path/to/ksail-go                # Navigate to repository root
pwd                                 # Should show path ending in 'ksail-go'
ls go.mod                           # Should exist
```

### Error Handling Guide

**Common failures and solutions:**

| Error                        | Likely Cause             | Solution                            |
|------------------------------|--------------------------|-------------------------------------|
| "go.mod not found"           | Wrong directory          | `cd` to repository root             |
| "Go version too old"         | Go < 1.24.0              | Update Go to 1.24.0+                |
| Tests fail                   | Code changes broke tests | Fix failing tests before proceeding |
| Build fails                  | Missing dependencies     | Run `go mod tidy` then retry        |
| Linter fails                 | Code quality issues      | Review output, fix violations       |
| mega-linter network timeouts | External link checking   | Normal - not a code quality issue   |

### Linting Strategy and Expectations

**Primary Linting** (REQUIRED for CI consistency):

```bash
mega-linter-runner -f go            # ~11-12 minutes
```

- Configuration in `.mega-linter.yml` with `APPLY_FIXES: all`
- **Auto-fixes** formatting and style issues automatically
- **Network timeouts expected**: Link checking may fail on external URLs (not a code quality issue)
- **Security scanning included**: Comprehensive analysis takes time
- **Current state**: Passes successfully (golangci-lint disabled in mega-linter config)

**Supplementary Go Linting** (optional for faster feedback):

```bash
~/go/bin/golangci-lint run          # ~1m16s - shows 0 issues
~/go/bin/golangci-lint run --fast-only  # Quick feedback
```

- Configuration in `.golangci.yml` with comprehensive rules
- **Current state**: Clean codebase, 0 issues
- **When to use**: During development for quick Go-specific feedback

**Linting Expectations:**

- **CI requirement**: Must use `mega-linter-runner -f go` for CI consistency per CONTRIBUTING.md
- **New code**: Should not introduce additional violations in enabled linters
- **Network failures**: Mega-linter link checking failures are not code quality issues
- **Performance**: Use golangci-lint for faster iteration during development

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

### Making Changes (Step-by-Step Process)

**ALWAYS follow this sequence when making code changes:**

1. **Validate current state**

   ```bash
   go test -v ./...                 # Must pass before starting
   ```

2. **Make your changes**
   - Edit code files as needed
   - Follow existing patterns (see [Codebase Navigation](#codebase-navigation))

3. **Generate mocks** (if interfaces changed)

   ```bash
   mockery                          # Regenerate mocks (~1.3s)
   ```

4. **Test changes**

   ```bash
   go test -v ./...                 # Must pass before proceeding
   ```

5. **Build verification**

   ```bash
   go build -o ksail .              # Must build successfully
   ./ksail --help                   # Basic functionality check
   ```

6. **Lint** (choose based on available time)

   ```bash
   # Full lint (CI-consistent):
   mega-linter-runner -f go         # ~11-12 minutes
   
   # OR quick Go-specific feedback:
   ~/go/bin/golangci-lint run       # ~1m16s
   ```

7. **Final validation**

   ```bash
   # Test complete CLI workflow
   ./ksail init --distribution Kind
   ./ksail up && ./ksail status && ./ksail list && ./ksail down
   ```

**Pre-commit hooks**: Hooks run automatically and may take extra time. Wait for completion - don't assume immediate success.

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

### Common Issues and Solutions

| Problem                  | Symptoms                            | Solution                                         |
|--------------------------|-------------------------------------|--------------------------------------------------|
| **Wrong directory**      | "go.mod not found"                  | `cd` to repository root (contains go.mod)        |
| **Old Go version**       | "Go directive...requires Go 1.24"   | Update Go to 1.24.0+                             |
| **Missing dependencies** | Build/import errors                 | Run `go mod download` then `go mod tidy`         |
| **Test failures**        | `go test` shows failures            | Fix failing tests before proceeding with changes |
| **Mock outdated**        | Interface-related test failures     | Run `mockery` to regenerate mocks                |
| **Linter fails**         | mega-linter or golangci-lint errors | Review linter output, fix code quality issues    |
| **Network timeouts**     | mega-linter link checking fails     | Normal behavior - not a code quality issue       |
| **CLI build issues**     | `./ksail` command not found         | Run `go build -o ksail .` to rebuild             |

### Debugging Steps

**If build fails:**

```bash
go version                          # Check Go version (need 1.24.0+)
go mod download                     # Ensure dependencies available
go mod tidy                         # Clean up go.mod/go.sum
go build -o ksail .                 # Retry build
```

**If tests fail:**

```bash
go test -v ./...                    # See detailed failure output
mockery                             # Regenerate mocks if interface changed
go test -v ./cmd/...                # Test specific package
UPDATE_SNAPSHOTS=true go test ./cmd/...  # Update snapshots if CLI output changed
```

**If CLI doesn't work:**

```bash
ls -la ksail                        # Check if binary exists
./ksail --help                      # Test basic functionality
./ksail --version                   # Check version output
```

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

**Last Updated**: Based on current repository state as of Go 1.25.1, CLI stub implementation. Validated on GitHub Actions environment with Docker 28.0.4, Kind v0.30.0, kubectl v1.34.0. All timings measured and commands tested to ensure accuracy of instructions. Updated to include comprehensive validation of all commands and current golangci-lint status (0 issues). Build times updated based on real measurements: ~0.5s first time, ~0.1s when cached.
