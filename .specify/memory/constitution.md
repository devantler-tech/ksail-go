<!--
Sync Impact Report
Version change: (none) → 1.0.0
Modified principles: (initial population)
Added sections: Engineering Standards & Constraints; Development Workflow & Review Gates; Governance (populated)
Removed sections: None
Templates requiring updates:
	- .specify/templates/plan-template.md ✅ updated (added Constitution Gate list)
	- .specify/templates/tasks-template.md ✅ updated (tests marked mandatory per Principle 2)
	- .specify/templates/spec-template.md ✅ aligned (already enforces independently testable stories)
	- .specify/templates/agent-file-template.md ✅ no change needed (descriptive aggregation only)
	- .specify/templates/checklist-template.md ✅ no change needed
Deferred TODOs:
	- RATIFICATION_DATE original adoption unknown → TODO(RATIFICATION_DATE) left for project maintainer to supply.
-->

# KSail-Go Constitution

## Core Principles

### I. Simplicity & Minimalism

Code MUST implement the simplest solution that fully satisfies current, approved
requirements. Avoid speculative abstractions (YAGNI). Functions SHOULD NOT
exceed ~50 logical lines; when they do, they MUST be refactored unless a clear
cohesive reason is documented in a plan Complexity Tracking table. Duplication
MUST be eliminated when it represents shared intent (DRY) but MAY be retained
when it improves clarity and reduces premature abstraction. Complexity
introductions (extra layers, patterns) REQUIRE justification in plan.md under
"Complexity Tracking" referencing the concrete problem solved.

### II. Test-First Quality Gates (NON-NEGOTIABLE)

All production logic MUST follow Red → Green → Refactor. For each new public
type, interface, or behavior: write failing unit/contract tests first, verify
failure, then implement code to pass. Mock generation via `mockery` MUST occur
before implementing dependent logic. A PR introducing public API without tests
is rejected. Minimum coverage threshold per package SHOULD be ≥80%; critical
packages (provisioner, installer, validator) MUST include contract tests for
external tool interactions. System/e2e tests run in CI matrices; local absence
is acceptable but MUST not break CI. Refactors MUST preserve existing tests or
update them transparently with rationale.

### III. Interface & Abstraction Discipline

High-level code MUST depend on small, focused interfaces (≤5 methods when
practical). Prefer composition over inheritance; avoid type switches that
bypass polymorphism. Implement SOLID: single responsibility per package and
struct; open for extension via new implementations, closed for modification of
stable contracts. Dependency inversion: constructors accept interfaces, not
concrete implementations. New providers (e.g., distribution, CNI installer)
MUST satisfy existing interfaces without requiring interface changes; if
changes are unavoidable they trigger a semantic version review.

### IV. Observability & Deterministic CLI Output

Every CLI command MUST produce structured, human-readable success output and
write errors exclusively to stderr. Timing instrumentation via the timer
package MUST appear in success messages (<1ms overhead). Logging MUST be
structured (key=value or JSON for machine output modes) and avoid sensitive
data. Commands offering machine-consumable output MUST provide a flag (e.g.
`--output json`) with deterministic field names. Silent failures are forbidden.

### V. Semantic Versioning & Backward Compatibility

Releases follow semantic versioning: MAJOR for breaking public API or CLI
behavior; MINOR for additive non-breaking features; PATCH for internal fixes,
docs, or non-functional improvements. Commit prefixes: `feat:` → MINOR, `fix:`
→ PATCH, `BREAKING CHANGE:` in body → MAJOR. Any removal or incompatible change
MUST include migration notes (added to release notes) and justification in the
PR description referencing affected interfaces. Avoid unnecessary breaking
changes by deprecating first when feasible.

## Engineering Standards & Constraints

- Language: Go ≥1.24.x (align with `go.mod`).
- Formatting: `go fmt` & `golangci-lint` MUST pass before merge; auto-fixes applied via pre-commit.
- Performance: Common CLI help commands SHOULD complete <200ms cold invocation. Timer overhead MUST remain <1ms cumulative per command.
- Resource Constraints: Avoid global mutable state; configuration passed via explicit structs.
- Security: No secrets in logs; error messages MUST avoid leaking credentials.
- Documentation: Public packages MUST have package comments; exported types MUST have doc comments meaningful to pkg.go.dev consumers.
- Code Smells: Long methods, large structs, primitive obsession, data clumps MUST trigger refactor tasks unless justified under Complexity Tracking.

## Development Workflow & Review Gates

- Feature proposal → `spec.md` with independently testable user stories.
- Planning → `plan.md` enumerates Constitution Check gates and Complexity Tracking justifications.
- Tasks → `tasks.md` groups work by user story; test tasks precede impl tasks.
- Implementation → Tests written & failing → code → refactor.
- Review Gate Checklist per PR:
  - Simplicity: Added abstractions justified? (Principle I)
  - Tests: New public APIs covered? (Principle II)
  - Interfaces: No oversized interfaces or type switches? (Principle III)
  - Observability: Timing & structured logging present? (Principle IV)
  - Versioning: Commit prefix aligns with change impact? (Principle V)
- Documentation updates required when CLI surface or public APIs change.
- Lint/test pipelines MUST be green; failing system tests block merge.

## Governance

This Constitution supersedes conflicting informal practices. Amendments follow:

1. Open a PR titled `governance: <summary>` including proposed diffs & impact analysis.
2. Classify version bump (MAJOR/MINOR/PATCH) with rationale referencing Principles or Workflow changes.
3. Provide migration guidance for any breaking change.
4. Obtain approval from at least one maintainer with governance mandate.
5. Merge auto-publishes new version (via semantic release) and updates `Last Amended` date.

Compliance Reviews: Quarterly (Jan/Apr/Jul/Oct) a maintainer audits packages
against Principles I–V and files issues for violations. Urgent violations may
trigger PATCH releases if fixes are non-breaking.

Enforcement: Reviewers MUST block merges that violate non-negotiable aspects:
Test-first (Principle II), structured error handling (Principle IV), semantic
version discipline (Principle V). Exceptions require explicit, time-bound
justification documented in the PR.

**Version**: 1.0.0 | **Ratified**: TODO(RATIFICATION_DATE): Original adoption date unknown | **Last Amended**: 2025-11-15
