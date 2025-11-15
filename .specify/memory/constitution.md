<!--
Sync Impact Report
Version change: (template) → 1.0.0
Modified principles: Initial population (none renamed)
Added sections: Engineering Standards & Constraints; Development Workflow & Review Gates; Governance (content)
Removed sections: None
Templates requiring updates:
	- .specify/templates/plan-template.md ✅ (add Constitution Check gates)
	- .specify/templates/tasks-template.md ✅ (tests mandatory per Principle II)
	- .specify/templates/spec-template.md ✅ (already aligns; no change)
	- .specify/templates/agent-file-template.md ✅ (no change needed)
	- .specify/templates/checklist-template.md ✅ (no change needed)
Follow-up TODOs: None
-->

# KSail-Go Constitution

## Core Principles

### I. Simplicity & Minimalism

Code MUST implement the simplest solution that fully satisfies approved
requirements. Avoid speculative abstractions (YAGNI). Functions SHOULD NOT
exceed ~50 logical lines; refactor when they do unless a cohesive reason is
documented in plan.md Complexity Tracking. Remove duplication that hides shared
intent (DRY); tolerate repetition that preserves clarity and avoids premature
abstraction. Added layers/patterns REQUIRE justification.

### II. Test-First Quality Gates (NON-NEGOTIABLE)

All production logic MUST follow Red → Green → Refactor. For each new public
type, interface, or behavior: write failing unit/contract tests first, verify
failure, then implement. Mock generation via `mockery` MUST precede dependent
logic. A PR adding public API without tests is rejected. Package coverage SHOULD
be ≥80%; critical packages (provisioner, installer, validator) MUST include
contract tests. Refactors preserve or transparently adapt existing tests.

### III. Interface & Abstraction Discipline

High-level code MUST depend on small, focused interfaces (≤5 methods when
practical). Prefer composition over inheritance; avoid type switches over
implementations. Apply SOLID: single responsibility per package/struct; extend
behavior via new implementations instead of modifying stable contracts. Use
dependency inversion: constructors accept interfaces—not concrete types. New
providers MUST implement existing interfaces; unavoidable interface changes
trigger semantic version review.

### IV. Observability & Deterministic CLI Output

Each CLI command MUST emit structured human-readable success output; errors go
only to stderr. Timing instrumentation (timer package) included in success
messages (<1ms overhead). Logging MUST be structured (key=value or JSON when
`--output json`). Machine output MUST be deterministic with stable field names.
No silent failures.

### V. Semantic Versioning & Backward Compatibility

Releases follow semver: MAJOR for breaking public API/CLI changes; MINOR for
additive non-breaking features; PATCH for internal fixes, docs, or non-functional
improvements. Commit prefixes: `feat:` → MINOR, `fix:` → PATCH, `BREAKING CHANGE:`
→ MAJOR. Removal/incompatible change MUST include migration notes in release
and PR description. Prefer deprecation before deletion.

## Engineering Standards & Constraints

- Language: Go ≥1.24.x (per go.mod).
- Formatting: `go fmt` & `golangci-lint` MUST pass before merge; auto-fixes via pre-commit.
- Performance: Common help commands SHOULD finish <200ms cold; timer overhead <1ms cumulative.
- State: Avoid global mutable state; pass configuration via structs.
- Security: No secrets in logs; error messages avoid credential leakage.
- Documentation: Public packages & exported types MUST have meaningful comments.
- Code Smells: Long methods, large structs, primitive obsession, data clumps MUST be refactored or justified.

## Development Workflow & Review Gates

- Feature proposal → `spec.md` with independently testable user stories.
- Planning → `plan.md` enumerates Constitution Check gates + Complexity Tracking.
- Tasks → `tasks.md` groups work by user story; test tasks precede implementation.
- Implementation → Write failing tests → implement → refactor.
- Review Gate Checklist per PR:
  - Simplicity (Principle I)
  - Tests present (Principle II)
  - Interface discipline (Principle III)
  - Observability/timing/logging (Principle IV)
  - Versioning alignment (Principle V)
- Documentation updates required for CLI/public API changes.
- Lint and test pipelines MUST be green; system test failures block merge.

## Governance

This Constitution supersedes conflicting informal practices.

Amendment Procedure:

1. PR titled `governance: <summary>` includes diffs + impact analysis.
2. Classify version bump (MAJOR/MINOR/PATCH) with rationale citing affected Principles/Workflow.
3. Provide migration guidance for any breaking change.
4. Obtain approval from at least one governance maintainer.
5. Merge updates `Last Amended` date and releases per semantic rules.

Compliance Reviews: Quarterly (Jan/Apr/Jul/Oct) audit of packages against Principles I–V; violations become issues. Urgent non-breaking fixes may trigger PATCH releases.

Enforcement: Reviewers MUST block merges violating non-negotiables: Test-first (II), structured error handling/logging (IV), semantic version discipline (V). Exceptions require time-bound documented justification.

**Version**: 1.0.0 | **Ratified**: 2025-11-15 | **Last Amended**: 2025-11-15
