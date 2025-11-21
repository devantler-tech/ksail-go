# Implementation Plan: Move All Go Source Code to src/

**Branch**: `001-move-all-source` | **Date**: 2025-11-16 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-move-all-source/spec.md`

> **⚠️ IMPLEMENTATION STATUS: NOT IMPLEMENTED**
>
> This plan was developed but **NEVER EXECUTED**. The codebase still has all Go source code at the repository root, not under a `src/` directory. This document represents planning work only.

## Summary

Reorganize the KSail-Go repository by moving all Go source code files, including `go.mod` and `go.sum`, from the repository root to a new `src/` subdirectory. This structural change maintains backward compatibility for external package consumers while establishing a clearer separation between source code and repository configuration files. The migration will be executed atomically in a single pull request with comprehensive pre-merge and post-merge validation.

## Technical Context

**Language/Version**: Go 1.25.4
**Primary Dependencies**: Cobra (CLI framework), Kind/K3d/eksctl (cluster provisioners), Kubernetes client-go, Flux CD APIs
**Storage**: File system (configuration files, cluster state)
**Testing**: go test (unit/integration), mockery (test mocks), system tests in CI
**Target Platform**: Cross-platform CLI (Linux, macOS, Windows)
**Project Type**: Single project - CLI application with package-based architecture
**Performance Goals**: Build times must remain unchanged (<2s for main build, <40s for full test suite)
**Constraints**:

- Zero breaking changes to external import paths (`github.com/devantler-tech/ksail-go/pkg/...`)
- IDE tooling must work after workspace reload
- CI/CD pipelines must pass with only path configuration updates

**Scale/Scope**: ~100+ Go source files across cmd/, pkg/, internal/ directories, 4 VS Code tasks, 3 GitHub Actions workflows

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Initial Check (Pre-Phase 0)

**Simplicity (I)**: ✅ **PASS** - This is a pure file reorganization without new abstractions, layers, or complex logic. All changes are structural file moves and configuration path updates.

**Test-First (II)**: ✅ **PASS** - No new public APIs are being introduced. Existing tests validate that functionality remains unchanged after the move. Validation strategy includes running full test suite pre-merge and post-merge.

**Interface Discipline (III)**: ✅ **PASS** - No new interfaces created. This is a structural change only, maintaining all existing interfaces.

**Observability (IV)**: ✅ **PASS** - No CLI commands are impacted functionally. Build and test commands will reference new paths, but command behavior and output remain identical. Error pathways unchanged - failures during migration trigger git revert.

**Versioning (V)**: **PATCH** - This is an internal structural change with zero external API impact. External consumers continue using the same import paths (`github.com/devantler-tech/ksail-go/pkg/...`). Module path in go.mod remains unchanged. No migration required for consumers.

### Post-Phase 1 Re-check

**Simplicity (I)**: ✅ **PASS** - Phase 1 design confirms pure structural change. Data model includes only three simple entities (FileSystemEntry, ConfigurationReference, ValidationCheckpoint) tracking the reorganization workflow. No new abstractions or complex patterns introduced.

**Test-First (II)**: ✅ **PASS** - Validation contracts define comprehensive pre-move, post-move, pre-merge, and post-merge checkpoints. All existing tests must pass at every checkpoint. No new test-first contracts required since no new public APIs are introduced.

**Interface Discipline (III)**: ✅ **PASS** - Phase 1 design confirms no new interfaces. All existing package interfaces remain unchanged. Import paths for external consumers unchanged.

**Observability (IV)**: ✅ **PASS** - Validation contracts include explicit build time tracking (baseline vs post-move) and rollback procedures. All validation checkpoints produce clear pass/fail signals with captured metrics.

**Versioning (V)**: **PATCH** - Phase 1 design confirms internal-only changes. Quickstart guide documents that external consumers experience zero breaking changes. Module path in go.mod remains `github.com/devantler-tech/ksail-go` throughout.

**Verdict**: All gates pass. Phase 1 design maintains constitutional compliance. Proceed to Phase 2 task breakdown.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

**Current Structure** (before reorganization):

```text
/
├── .github/             # CI/CD workflows, scripts
├── .vscode/             # VS Code workspace config
├── cmd/                 # CLI commands (Cobra)
├── pkg/                 # Core business logic packages
├── internal/            # Internal utilities
├── main.go              # Application entry point
├── main_test.go         # Main package tests
├── go.mod               # Go module definition
├── go.sum               # Go module checksums
├── bin/                 # Built binaries output
├── docs/                # Documentation
├── k8s/                 # Sample K8s manifests
├── schemas/             # JSON schemas
├── specs/               # Feature specifications
└── [config files]       # .golangci.yml, .mockery.yml, kind.yaml, etc.
```

**Target Structure** (after reorganization):

```text
/
├── .github/             # CI/CD workflows, scripts (STAYS)
├── .vscode/             # VS Code workspace config (STAYS)
├── src/                 # ✨ NEW: All Go source code
│   ├── cmd/             # CLI commands (MOVED)
│   ├── pkg/             # Core packages (MOVED)
│   ├── internal/        # Internal utilities (MOVED)
│   ├── main.go          # Entry point (MOVED)
│   ├── main_test.go     # Main tests (MOVED)
│   ├── go.mod           # Module definition (MOVED)
│   └── go.sum           # Module checksums (MOVED)
├── bin/                 # Built binaries output (STAYS or reconfigured)
├── docs/                # Documentation (STAYS)
├── k8s/                 # Sample K8s manifests (STAYS)
├── schemas/             # JSON schemas (STAYS)
├── specs/               # Feature specifications (STAYS)
└── [config files]       # Project configs (STAY at root)
```

**Structure Decision**: Single CLI project with clean separation between source code (src/) and repository configuration. The Go module root moves to src/, establishing it as the working directory for all Go commands. This maintains the existing package-based architecture (cmd/, pkg/, internal/) while providing clearer repository organization.

## Complexity Tracking

**Status**: ✅ **NONE** - All constitutional gates pass. This is a straightforward structural reorganization with no complexity violations.
