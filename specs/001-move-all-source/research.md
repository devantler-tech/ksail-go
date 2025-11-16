# Research: Move All Go Source Code to src/

**Feature**: Move All Go Source Code to src/
**Phase**: 0 - Research & Discovery
**Date**: 2025-11-16

## Research Questions

### 1. Go Module Support for Subdirectory Modules

**Question**: Does Go tooling fully support having the module root (go.mod) in a subdirectory rather than the repository root?

**Research Findings**:

- **Decision**: Yes, Go fully supports modules in subdirectories
- **Rationale**: Go modules are defined by the location of go.mod, not by repository structure. The module path in go.mod remains `github.com/devantler-tech/ksail-go`, and the physical location of go.mod can be anywhere in the repository
- **Impact**: Commands must be run from src/ directory or use `-C src` flag (Go 1.20+)
- **Alternatives Considered**:
  - Keep go.mod at root: Rejected because it defeats the purpose of organizing all Go source in src/
  - Use replace directives: Rejected as unnecessary complexity for a simple directory move

### 2. Import Path Compatibility

**Question**: Will external consumers' import paths continue to work when go.mod moves to src/?

**Research Findings**:

- **Decision**: Yes, import paths remain fully compatible
- **Rationale**: Go import paths are based on the module path declared in go.mod (`module github.com/devantler-tech/ksail-go`), not the physical location of go.mod in the repository. External consumers will see zero changes
- **Evidence**: The module path stays identical, and Go's module system resolves imports based on this declaration
- **Alternatives Considered**: None needed - this is a non-issue

### 3. Build Command Patterns

**Question**: What are the best practices for running Go commands with a subdirectory module?

**Research Findings**:

- **Decision**: Use `cd src && go <command>` or `go -C src <command>` (Go 1.20+)
- **Rationale**: Go 1.20 introduced the `-C` flag for changing directories before executing commands, providing flexibility for both local development and CI/CD
- **Best Practices**:
  - CI/CD workflows: Set `working-directory: src` in GitHub Actions
  - VS Code tasks: Update `cwd` to `${workspaceFolder}/src`
  - Local development: Either cd to src/ or use `-C src` flag
  - Scripts: Add `cd src` at the beginning or use `-C src` flag
- **Alternatives Considered**:
  - Wrapper scripts: Rejected as adding unnecessary indirection
  - Symbolic links: Rejected as platform-specific and maintenance burden

### 4. IDE Integration

**Question**: How do popular Go IDEs handle modules in subdirectories?

**Research Findings**:

- **Decision**: All major IDEs support subdirectory modules with proper configuration
- **VS Code**: Go extension automatically detects go.mod location; may require workspace reload after structural changes
- **GoLand/IntelliJ**: Automatically recognizes Go modules in subdirectories; project reload recommended
- **Vim/Neovim with gopls**: Language server finds go.mod automatically in parent directories
- **Best Practice**: Document that developers should reload workspace/restart IDE after pulling the reorganization changes
- **Alternatives Considered**: None needed - IDE support is mature

### 5. Version Control History Preservation

**Question**: How can we preserve git history and blame information when moving files?

**Research Findings**:

- **Decision**: Use `git mv` command for all file moves
- **Rationale**: `git mv` explicitly tracks file moves, preserving history. Git's rename detection also works retrospectively with `git log --follow` and `git blame -C`
- **Best Practices**:
  - Use `git mv` for each file/directory move
  - Commit file moves separately from other changes for cleaner history
  - Document the reorganization in commit message for future reference
- **Alternatives Considered**:
  - Manual move + git add: Rejected because it's less explicit about preserving history
  - Rewrite history: Rejected as unnecessary and disruptive

### 6. CI/CD Pipeline Updates

**Question**: What specific changes are needed in GitHub Actions workflows?

**Research Findings**:

- **Decision**: Update working-directory and command paths in all workflow steps
- **Required Changes**:
  - Add `working-directory: src` to Go-related steps
  - Update paths in GoReleaser configuration to reference src/main.go
  - Update schema generation scripts to use src/ paths
  - Update mockery configuration if it references absolute paths
- **Files to Update**:
  - `.github/workflows/ci.yaml`: Add working-directory to build/test/lint steps
  - `.github/workflows/cd.yaml`: Update working-directory for release builds
  - `.github/workflows/release.yaml`: Update GoReleaser paths
  - `.github/scripts/generate-schema.sh`: Update source paths
- **Alternatives Considered**: None - direct path updates are simplest

### 7. Binary Output Configuration

**Question**: Where should compiled binaries be output after the reorganization?

**Research Findings**:

- **Decision**: Keep bin/ at repository root, configure output path explicitly
- **Rationale**: Binaries are build artifacts separate from source code. Keeping them at root maintains consistency with repository patterns
- **Implementation**: Use `go build -o ../bin/ksail` when building from src/
- **VS Code Tasks**: Update build task to use correct output path
- **GoReleaser**: Update binary output paths in .goreleaser.yaml
- **Alternatives Considered**:
  - Move bin/ to src/: Rejected because build artifacts shouldn't be in source directory
  - Output to src/bin/: Rejected for same reason

### 8. Configuration File Locations

**Question**: Which configuration files should move to src/ and which should stay at root?

**Research Findings**:

- **Decision**: Only Go source and go.mod/go.sum move to src/; all config files stay at root
- **Stay at Root**:
  - `.golangci.yml` (linter config - repository-level)
  - `.mockery.yml` (mock generator config - may need path updates)
  - `.github/` (CI/CD workflows)
  - `.vscode/` (workspace settings - may need path updates)
  - `kind.yaml`, `k3d.yaml`, `ksail.yaml` (sample configs)
  - All other dotfiles (.gitignore, .pre-commit-config.yaml, etc.)
- **Move to src/**:
  - All .go files (cmd/, pkg/, internal/, main.go, main_test.go)
  - go.mod and go.sum
- **Rationale**: Separation of source code from repository/project configuration
- **Alternatives Considered**: Moving all config to src/ - rejected as reducing repository-level visibility

## Summary of Decisions

| Topic | Decision | Rationale |
|-------|----------|-----------|
| Module location | go.mod in src/ | Go fully supports subdirectory modules |
| Import paths | No changes needed | Module path declaration is what matters |
| Build commands | `cd src && go <cmd>` or `go -C src <cmd>` | Standard Go practices for subdirectory modules |
| IDE support | Workspace reload required | All major IDEs support with reload |
| Git history | Use `git mv` | Preserves history and blame information |
| CI/CD updates | Add working-directory: src | Clean and explicit configuration |
| Binary output | Keep bin/ at root | Separate artifacts from source |
| Config files | Stay at root (except go source) | Repository-level configuration visibility |

## Implementation Readiness

✅ All research questions resolved with clear decisions
✅ No blocking technical issues identified
✅ Best practices documented for each concern area
✅ Ready to proceed to Phase 1 (Design)
