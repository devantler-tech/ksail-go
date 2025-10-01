<!--
SYNC IMPACT REPORT
==================
Version Change: None → 1.0.0 (Initial Constitution)
Modified Principles: N/A (initial creation)
Added Sections:
  - All core principles (I-VII)
  - Architecture Standards
  - Quality Gates & Development Workflow
  - Governance
Templates Requiring Updates:
  ✅ plan-template.md - Reviewed, Constitution Check section compatible
  ✅ spec-template.md - Reviewed, requirement alignment verified
  ✅ tasks-template.md - Reviewed, task categorization aligned with principles
Follow-up TODOs: None
-->

# KSail Constitution

## Core Principles

### I. Package-First Design

Every feature MUST start as a standalone package in `pkg/` before any CLI implementation. Packages MUST be:

- Self-contained with clear domain boundaries
- Independently testable without external dependencies
- Designed as public API that external applications can import
- Documented with comprehensive README files and GoDoc comments

**Rationale**: KSail is both a CLI tool and a reusable SDK. Package-first design ensures features are composable, testable in isolation, and valuable beyond the CLI context. This architectural pattern enforces separation of concerns and enables external developers to build on KSail's functionality.

### II. CLI Interface

Every package MUST expose its functionality through CLI commands using the Cobra framework. Commands MUST:

- Accept input via flags, arguments, or stdin
- Write success output to stdout and errors to stderr
- Support both human-readable and structured output formats where applicable
- Follow consistent command hierarchy (`cluster`, `workload` namespaces)
- Provide comprehensive help text and usage examples

**Rationale**: The CLI is the primary user interface for KSail. Consistent command patterns improve discoverability and usability. Structured output enables scripting and automation workflows.

### III. Test-First Development (NON-NEGOTIABLE)

TDD is MANDATORY for all code changes. The workflow MUST be:

1. Write failing tests (unit, integration, or contract tests)
2. Run tests to verify they fail for the right reason
3. Implement minimal code to make tests pass
4. Refactor with tests as safety net
5. Generate mocks via `mockery` before implementation

Tests MUST be written before implementation code in every PR. No exceptions.

**Rationale**: TDD ensures testability, catches regressions early, and documents expected behavior. The mockery tool eliminates boilerplate and enforces interface-based design. This principle is non-negotiable because retrospective testing often results in untestable code and inadequate coverage.

### IV. Interface-Based Design

All core functionality MUST be defined as interfaces before implementation. Implementations MUST:

- Accept dependencies as interface parameters (dependency injection)
- Use Go context for cancellation and timeouts on long-running operations
- Be mockable via the mockery tool
- Have clear contracts documented in interface definitions

**Rationale**: Interface-based design enables testing, extensibility, and loose coupling. It allows different implementations (e.g., Kind, K3d, EKS provisioners) to conform to common contracts. Mockery-generated mocks ensure consistent test doubles across the codebase.

### V. Clean Architecture Principles

Code MUST follow clean architecture patterns:

- **Domain Separation**: Each package addresses a specific domain (provisioner, config-manager, installer, etc.)
- **Dependency Direction**: Dependencies point inward toward domain logic, never outward
- **No Circular Dependencies**: Package imports form a directed acyclic graph
- **Context Propagation**: Use `context.Context` for cancellation, timeouts, and request-scoped values

**Rationale**: Clean architecture ensures maintainability, testability, and evolvability. Domain separation prevents tangled dependencies. Context propagation enables graceful shutdown and resource cleanup.

### VI. Quality Gates

All code changes MUST pass these quality gates before merge:

1. **Lint Pass**: `golangci-lint run` (timeout: 90+ seconds)
2. **Unit Tests**: `go test ./...` (timeout: 60+ seconds)
3. **Build Success**: `go build ./...`
4. **Mock Generation**: `mockery` generates up-to-date mocks
5. **Pre-commit Hooks**: Automated via pre-commit framework
6. **System Tests**: Full lifecycle validation (CI only)

Mega-linter validation SHOULD be run periodically for comprehensive validation.

**Rationale**: Automated quality gates catch issues before review, reduce cognitive load on reviewers, and maintain consistent code quality. The specified timeouts prevent premature cancellation of long-running validation tasks.

### VII. Semantic Versioning & Conventional Commits

Version management MUST follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking API changes, removed functionality
- **MINOR**: New features, backward-compatible additions
- **PATCH**: Bug fixes, documentation, non-breaking refactors

Releases are automated via PR titles:

- `fix:` triggers PATCH release
- `feat:` triggers MINOR release
- `BREAKING CHANGE:` or `!` suffix triggers MAJOR release

**Rationale**: Semantic versioning communicates impact to users. Conventional commits automate release management and changelog generation. This eliminates manual versioning errors and ensures clear communication of changes.

## Architecture Standards

### Repository Structure

The project MUST maintain this structure:

```text
cmd/                    # CLI command implementations (Cobra)
pkg/                    # Core business logic (public API)
  ├── apis/            # Kubernetes API definitions
  ├── config-manager/   # Configuration management
  ├── installer/        # Component installers
  ├── io/              # Safe file I/O operations
  ├── provisioner/      # Cluster provisioning
  └── ...
internal/               # Private utilities (not importable)
specs/                  # Feature specifications and plans
.specify/               # Specification framework templates
```

### Technology Stack

**Required Dependencies**:

- Go 1.24.0+ (language version)
- Cobra (CLI framework)
- mockery v3.x (mock generation)
- golangci-lint (code quality)
- Docker (cluster provisioning)

**Core Libraries**:

- `sigs.k8s.io/kind` (Kind cluster management)
- `github.com/k3d-io/k3d/v5` (K3d cluster management)
- `github.com/weaveworks/eksctl` (EKS cluster management)
- `k8s.io/client-go` (Kubernetes client)

## Quality Gates & Development Workflow

### Pre-Commit Validation

Developers MUST run these commands before committing:

```bash
mockery                    # Generate fresh mocks
go test ./...             # Run all tests (~37s)
golangci-lint run         # Lint code (~1m16s)
go build -o ksail .       # Ensure clean build
```

Pre-commit hooks automate formatting and mock generation.

### Manual Testing Requirements

After code changes affecting CLI or cluster operations, developers MUST validate:

1. **CLI Help**: Verify `--help` output for affected commands
2. **Basic Lifecycle**: Test `init → up → status → down` workflow
3. **Distribution Support**: Validate changes against Kind (required), K3d/EKS (if affected)

### CI Pipeline Requirements

The CI pipeline enforces:

- Standard Go CI (build, test, lint) via reusable workflows
- System tests for each distribution (Kind, K3d, EKS)
- Full lifecycle validation (complete workflow from init to down)
- Code coverage reporting via Codecov

### Code Review Standards

Reviewers MUST verify:

- Tests are written before implementation (TDD compliance)
- Interfaces are defined before concrete implementations
- Mock generation is up-to-date
- Changes align with package-first design
- No violations of clean architecture principles
- Quality gates pass in CI

## Governance

This constitution supersedes all other development practices and guidelines. When conflicts arise between this document and other documentation, this constitution takes precedence.

**Amendment Process**:

1. Propose amendment via issue or PR with clear rationale
2. Document impact on existing code and templates
3. Update affected templates (plan-template.md, spec-template.md, tasks-template.md)
4. Increment constitution version according to semantic versioning:
   - **MAJOR**: Backward-incompatible governance changes, principle removals
   - **MINOR**: New principles added, materially expanded guidance
   - **PATCH**: Clarifications, wording improvements, non-semantic refinements
5. Update Last Amended date

**Compliance Review**:

- All PRs MUST demonstrate constitutional compliance
- Feature specs MUST include Constitution Check section
- Implementation plans MUST verify compliance at Phase 0 and Phase 1 gates
- Violations MUST be justified or corrected before merge

**Development Guidance**:
Runtime development guidance is maintained in `.github/copilot-instructions.md`. This file provides AI coding assistants with project-specific context, timing expectations, and validation procedures.

**Version**: 1.0.0 | **Ratified**: 2025-10-01 | **Last Amended**: 2025-10-01
