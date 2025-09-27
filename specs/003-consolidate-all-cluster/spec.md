# Feature Specification: Consolidate Cluster Commands Under `ksail cluster`

**Feature Branch**: `003-consolidate-all-cluster`
**Created**: 2025-09-27
**Status**: Draft
**Input**: User description: "Consolidate all cluster related commands under a `ksail cluster x` command, where x is the existing cluster related commands.

This is important to ensure CLI UX stays intuitive and easy and uncluttered as ksail grows in features and CLI commands. It is important for users.

A full codebase already exists, so this is an enhancement of the existing codebase, and it is important the plan and tasks generated from this spec is based on the existing codebase."

## Clarifications

### Session 2025-09-27

- Q: What should happen when a user runs a legacy top-level command like `ksail up` after we move everything beneath `ksail cluster`? → A: Remove entirely (command not found)
- Q: How should we inform existing scripted or automated workflows about the command structure change so they can migrate smoothly? → A: No additional messaging beyond standard documentation updates
- Q: How should the top-level `ksail --help` output reflect the new `cluster` grouping so users discover it quickly? → A: List `cluster` among commands with a brief description only
- Q: What concrete signals should we require before calling this feature done? → A: Merge after code review; no extra validation

## User Scenarios & Testing *(mandatory)*

### Primary User Story

As a platform engineer using the KSail CLI, I want all cluster lifecycle commands grouped under `ksail cluster` so that managing environments feels organized and scalable as new features arrive.

### Acceptance Scenarios

1. **Given** an operator authenticated on their workstation, **When** they run `ksail cluster up`, **Then** the tool provisions a cluster exactly as the former `ksail up` command did and reports success or failure clearly.
2. **Given** an operator exploring available operations, **When** they run `ksail cluster --help`, **Then** the CLI lists all cluster subcommands with concise explanations so the user understands available actions.
3. **Given** an operator, **When** they run `ksail cluster reconcile`, **Then** the CLI reports an unknown command error, and `reconcile` remains available only at the top level.

### Edge Cases

- Legacy commands such as `ksail up` will be removed; invoking them must produce the standard unknown-command error so users adopt the new `ksail cluster` syntax.
- Scripted or automated workflows rely on standard release documentation updates; no extra CLI messaging is planned for the command structure change.
- Running `ksail cluster` without a subcommand must output the cluster help/usage so users immediately see available actions.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST expose a parent `cluster` command (invoked as `ksail cluster`) dedicated to cluster lifecycle actions.
- **FR-002**: The CLI MUST provide subcommands under `ksail cluster` for every existing cluster lifecycle capability (e.g., create, delete, start, stop, list, status) with the same user-facing behavior as today. (**Note:** `reconcile` is intentionally excluded and will be migrated separately.)
- **FR-003**: Executing any cluster subcommand via `ksail cluster <subcommand>` MUST produce identical success/failure handling, messaging, and exit codes as the current standalone commands.
- **FR-004**: Invoking `ksail cluster --help` MUST display purpose-driven guidance that orients users to the consolidated command structure.
- **FR-005**: Running `ksail cluster` without a subcommand MUST display the cluster help/usage output so the workflow remains discoverable.
- **FR-006**: Legacy top-level commands (e.g., `ksail up`, `ksail down`) MUST be removed so invoking them produces the CLI's standard unknown-command error, ensuring users transition to `ksail cluster <subcommand>` workflows.
- **FR-007**: The root `ksail --help` output MUST include a command list entry for `cluster` with a concise description, with no additional dedicated sections or migration notes.
- **FR-008**: The `reconcile` command MUST NOT be moved under `ksail cluster` in this refactor. It will remain at the top level until migrated to `ksail workloads reconcile` in a future change.

## Dependencies & Assumptions

- Existing Cobra command constructors (`NewUpCmd`, `NewDownCmd`, etc.) and handlers (`HandleUpRunE`, ...) remain stable and reusable under the new parent command without signature changes.
- Snapshot testing infrastructure (`go-snaps` fixtures under `cmd/__snapshots__`) is available so help output deltas can be captured once tests are updated.
- Build and smoke validation steps rely on the local `./ksail` binary produced via `go build ./...`.
- Constitution quality gates (golangci-lint, coverage, performance thresholds) stay enforced by existing CI pipelines; no new tooling is required.

## Future Work / Out of Scope

- Migration of the `reconcile` command to `ksail workloads reconcile` will be handled in a future feature and is explicitly out of scope for this refactor.

## Completion Signals

Feature is considered complete only when all constitution-mandated quality gates are met:

- All code and tests pass golangci-lint with zero issues
- All tests are written before implementation (strict TDD)
- >90% test coverage is achieved and validated
- CLI response and performance meet defined thresholds
- Documentation is updated for all user-facing changes
- Manual and automated validation steps in tasks.md are complete

Merge is allowed only after code review and all above gates are satisfied.

---

## Review & Acceptance Checklist

GATE: Automated checks run during main() execution

### Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status

Updated by main() during processing

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified: CLI command tree (root, cluster parent, subcommands), help output, error messages, legacy command aliases
- [x] Review checklist passed

---
