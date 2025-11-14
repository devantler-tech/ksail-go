<!--
Sync Impact Report
Version: 0.0.0 → 1.0.0
Modified Principles:
- (new) I. Code Quality Discipline
- (new) II. Testing Rigor
- (new) III. User Experience Consistency
- (new) IV. Performance & Reliability Contracts
Added Sections:
- Engineering Guardrails
- Workflow & Review Gates
Templates Requiring Updates:
- .specify/templates/plan-template.md ✅ updated
- .specify/templates/spec-template.md ✅ updated
- .specify/templates/tasks-template.md ✅ updated
Deferred TODOs: none
-->

# KSail Go Constitution

## Core Principles

### I. Code Quality Discipline

- All Go source files MUST be formatted with `gofmt`/`goimports` and free of `golangci-lint` warnings before review.
- Designs MUST favor composition and small, testable units (functions ideally ≤ 50 lines) with clear boundaries between packages.
- Public CLI changes MUST include updated documentation or help text to prevent drift.

**Rationale**: Enforcing consistent style, structure, and documentation keeps the CLI maintainable as contributions scale.

### II. Testing Rigor

- Every behavioral change MUST ship with unit tests; integration/system tests are REQUIRED when touching provisioning, cluster lifecycle, or external clients.
- Tests MUST be deterministic, table-driven when asserting multiple cases, and isolate IO via Cobra streams—no direct `os.Stdout` writes.
- `go test ./...` MUST pass locally and in CI before merge.

**Rationale**: Repeatable tests catch regressions early and preserve trust in automation-heavy workflows.

### III. User Experience Consistency

- CLI commands MUST use the shared `notify` and `timer` utilities for messaging, success, and error flows; no ad-hoc `fmt.Println` output.
- Commands MUST respect Cobra-provided IO streams so piping and scripting behave consistently.
- User-facing changes MUST include usage examples or quickstart updates, ensuring parity between docs and behavior.

**Rationale**: Consistent messaging and documentation make the CLI predictable for both humans and automation.

### IV. Performance & Reliability Contracts

- Long-running operations MUST provide progress feedback and adhere to documented timeouts; new workflows MUST state expected runtime budgets.
- Resource-intensive features MUST include benchmarks or measurements when they risk exceeding baseline cluster start times.
- Any change that affects concurrency, retries, or backoff MUST document failure handling to avoid silent degradations.

**Rationale**: Explicit performance expectations protect developer productivity and cluster stability.

## Engineering Guardrails

- Language/runtime: Go 1.25.x with module consistency enforced via `go mod tidy` in CI.
- Tooling: `golangci-lint` (project config), `mockery` for interface mocks, Markdown linting, and lychee link checks.
- Secrets or credentials MUST flow through existing SOPS/AGE pipelines—no plain-text artifacts.
- External dependencies require pinned versions and release notes linked in PRs before adoption.

## Workflow & Review Gates

- Every `/speckit.plan` MUST pass the Constitution Check by confirming adherence to all four core principles.
- Specs MUST enumerate acceptance tests (Principle II) and note UX impact (Principle III) plus performance budgets (Principle IV).
- Task breakdowns MUST isolate work per user story, include explicit testing tasks, and flag any code-quality refactors.
- Pull requests MUST attach evidence (command output or CI links) for `go test ./...`, lint, and—when applicable—benchmarks.

## Governance

- **Authority**: This constitution supersedes prior undocumented conventions; conflicts defer to the most specific principle.
- **Amendments**: Proposals require an ADR or PR describing the change, its impact on templates, and migration steps.
- **Compliance**: Reviewers MUST block merges that violate any principle or guardrail until remediated.
- **Versioning**: Semantic increments—MAJOR for added/removed principles, MINOR for new guidance, PATCH for clarifications.

**Version**: 1.0.0 | **Ratified**: 2025-11-14 | **Last Amended**: 2025-11-14
