# KSail Go - Kubernetes Cluster Management CLI

KSail is a Go-based CLI tool for managing local Kubernetes clusters and workloads declaratively. **Currently a work-in-progress migration from a previous implementation.** All CLI subcommands are implemented as working stubs with proper help, flags, and basic functionality.

Always reference these instructions first and fallback to search or bash commands only when you encounter unexpected information that does not match the info here.

## Working Effectively

### Bootstrap and Build

- **Install Go 1.23.9+**: Verify with `go version` (go.mod requires 1.23.9, CD pipeline uses 1.25.0)
- **Download dependencies**: `go mod download` (completes in ~5 seconds)
- **Build the application**: `go build -o ksail .` -- takes ~16 seconds initially, <1 second when cached. Set timeout to 60+ seconds for safety.
- **Install golangci-lint v2**: `curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ~/go/bin latest`

### Required Dependencies

- **Go 1.23.9+**: Programming language runtime (required, go.mod specifies 1.23.9)
- **Docker**: Container runtime - verify with `docker --version` (available and used in system tests)
- **Kind**: Local Kubernetes clusters - verify with `kind version` (available and used in system tests)  
- **kubectl**: Kubernetes CLI - verify with `kubectl version --client` (available and used in system tests)

### Testing and Validation

- **Run tests**: `go test -v ./...` -- takes ~9 seconds. Set timeout to 60+ seconds for safety.
- **Run linter**: `~/go/bin/golangci-lint run` -- takes ~7 seconds. Set timeout to 60+ seconds for safety.
  - **Current State**: 0 issues (clean codebase), expected wsl deprecation warning (non-breaking)

### Application Usage

- **Get help**: `./ksail --help` (shows all subcommands)
- **Get version**: `./ksail --version` (shows dev version info)
- **Available Commands**: `init`, `up`, `down`, `start`, `stop`, `list`, `status`, `reconcile`, `completion`
- **Example**: `./ksail init --container-engine Docker --distribution Kind`

## Validation Scenarios

### CRITICAL: Always test complete workflows after making changes

1. **Build Validation**: `go build -o ksail .` → `./ksail --help` → `./ksail --version`
2. **Development Workflow**: Tests → Build → Basic functionality → Linting
3. **System Testing**: `./ksail init --container-engine Docker --distribution Kind` → `./ksail up` → `./ksail status` → `./ksail down`

## Build Timings

- **`go build -o ksail .`**: ~16s initially, <1s cached -- SET TIMEOUT TO 60+ SECONDS
- **`go test -v ./...`**: ~9s -- SET TIMEOUT TO 60+ SECONDS  
- **`~/go/bin/golangci-lint run`**: ~7s -- SET TIMEOUT TO 60+ SECONDS
- **`go mod download`**: ~5s -- SET TIMEOUT TO 60+ SECONDS

### Linting Expectations

- **Current State**: 0 linting issues (clean codebase)
- **Config File**: `.golangci.yml` (note: .yml extension, not .yaml)
- **Command**: `~/go/bin/golangci-lint run` (use full path, not system golangci-lint)
- **Focus**: Ensure new code doesn't introduce violations

## Codebase Navigation

### Directory Structure (~780 lines Go code)

- **`cmd/`**: CLI commands (root.go 84 lines, individual commands 25-33 lines each)
  - `ui/` - notify (120 lines), quiet (29 lines), asciiart (73 lines)
  - `__snapshots__/` - Snapshot test files
- **`internal/utils/path/`**: Path utilities (19 lines)
- **`pkg/provisioner/cluster/`**: Cluster logic (215 lines)
- **`main.go`**: Entry point (28 lines)

### Key Patterns

- **CLI Framework**: Cobra with consistent command structure
- **Testing**: go-snaps for snapshot testing, comprehensive coverage
- **UI**: fatih/color with symbols (✓, ✗, ⚠, ►)
- **Architecture**: Clean separation: cmd (interface), internal (utilities), pkg (business logic)

## Configuration Files

- **`.golangci.yml`**: v2 format with depguard restrictions. Allows: $gostd, project, docker, fatih/color, k3d, cobra, sigs.k8s.io
- **`go.mod`**: Go 1.23.9+ with Cobra, Docker SDK, fatih/color, go-snaps, uber/mock, kind
- **`.github/workflows/`**: CI (external + system testing matrix), CD (GoReleaser + Docker)
- **`.goreleaser.yaml`**: Multi-platform builds with version injection via ldflags
- **`Dockerfile`**: Distroless static image with health check

## Development Workflow

1. **Always validate current state**: `go test -v ./...`
2. **Build and test**: `go build -o ksail .` → `./ksail --help`
3. **Lint**: `~/go/bin/golangci-lint run`
4. **Follow patterns**: Study `cmd/root.go`, use Cobra framework, add tests with go-snaps

## Testing & Troubleshooting

### Testing Strategy
- **Unit Tests**: `*_test.go` files with go-snaps snapshot testing
- **System Tests**: CI matrix with Docker/Podman + Kind/K3d combinations
- **Coverage**: All CLI commands and UI components tested

### Common Issues
- **Build fails**: Check Go version (1.23.9+), run `go mod download`
- **Linter issues**: Use `~/go/bin/golangci-lint run`, ignore wsl deprecation warning
- **Import violations**: Check `.golangci.yml` depguard rules
- **Test failures**: Check `__snapshots__/` directories
- **Docker issues**: Verify Docker daemon running for system tests

## CI/CD

- **CI**: External reusable workflow + system testing matrix (Docker/Podman × Kind/K3d)
- **CD**: GoReleaser with multi-platform builds and Docker images on version tags
- **Quality Gates**: Comprehensive linting, testing, and system validation

---

**Repository Status**: ~780 lines Go code, comprehensive CLI with stub functionality, clean architecture, 0 linting issues
