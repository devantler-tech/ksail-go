# Tasks: Consolidate Cluster Commands Under `ksail cluster`

> **Scope Guard – `reconcile` remains top-level:** This refactor must not move `reconcile` under `ksail cluster`; keep it at the root until the future `ksail workloads reconcile` migration.

**Input**: Design documents from `/specs/003-consolidate-all-cluster/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Requirement Coverage Checklist

| Requirement Key | Description                                                           | Covered By Task(s)  |
| --------------- | --------------------------------------------------------------------- | ------------------- |
| FR-001          | Expose parent `cluster` command                                       | T001, T006          |
| FR-002          | Subcommands for all cluster lifecycle actions (excluding 'reconcile') | T004, T006, T007, T011 |
| FR-003          | Subcommand behavior/messaging parity                                  | T002, T003, T004, T007, T011, T012 |
| FR-004          | `ksail cluster --help` guidance                                       | T003, T005, T009, T012 |
| FR-005          | `ksail cluster` (bare) help output                                    | T002, T005, T012 |
| FR-006          | Remove legacy top-level commands                                      | T008, T012 |
| FR-007          | Root help lists `cluster`                                             | T003, T005, T009, T011, T012 |
| FR-008          | Exclude `reconcile` from cluster refactor                             | T014 |

*Non-functional gates (NFR-Performance, NFR-Lint, NFR-Coverage) are tracked via tasks T010 and T015–T017.*

## Non-Functional Validation

- [ ] T015 (NFR-Performance) Validate CLI response and performance meet defined thresholds
- [ ] T016 (NFR-Lint) Validate all code and tests pass golangci-lint with zero issues
- [ ] T017 (NFR-Coverage) Validate >90% test coverage is achieved and validated

## Phase 3.1: Setup

- [ ] T001 (FR-001) Prepare `cmd/cluster/` package scaffold (create directory and lightweight `doc.go` explaining the consolidated cluster command namespace) so subsequent tests can target the new package.

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

- [ ] T002 (FR-005, FR-003) Add explicit test for `ksail cluster` (bare) help output and ensure messaging parity with legacy commands (TDD) (i.e., verify that help text structure, error message format, and command output are consistent with the legacy commands)
- [ ] T003 (FR-004, FR-007) Extend root command help snapshots to assert `cluster` appears in `ksail --help` and that `cluster --help` includes all lifecycle verbs except `reconcile`.
- [ ] T004 (FR-002, FR-003, FR-004) Create table-driven tests that confirm each lifecycle subcommand is bound under `cluster` with the same short/long descriptions as the legacy top-level commands.
- [ ] T005 (FR-004, FR-005, FR-007) Update CLI UI/help fixtures so the quickstart snapshot and usage guidance cover `ksail cluster` invocation patterns before implementation begins.

## Phase 3.3: Implementation

- [ ] T006 (FR-001, FR-002) Introduce the parent `cluster` Cobra command in `cmd/cluster/` and register it on the root command with an accurate summary.
- [ ] T007 (FR-002, FR-003) Move existing lifecycle command constructors (`up`, `down`, `start`, `stop`, `status`, `list`) under the new parent while preserving command wiring and examples.
- [ ] T008 (FR-006) Remove legacy top-level lifecycle command registrations and adjust any remaining references to point at the new group.
- [ ] T009 (FR-004, FR-007) Update user-facing help text, usage examples, and quickstart documentation to reflect the grouped `cluster` commands and regenerate relevant markdown/JSON artifacts.

## Phase 3.4: Integration & Validation

- [ ] T010 (NFR-Lint) [P] Run `gofmt`/`goimports` on all touched Go files (including the new `cmd/cluster/` directory) to satisfy formatting gates. *(Depends on: T007)*
- [ ] T011 (FR-002, FR-003, FR-007, NFR-Performance, NFR-Coverage) Execute `go test ./cmd` (updating go-snaps snapshots as needed) followed by `go test ./...` to confirm all suites pass with the new command structure. *(Depends on: T010)*
- [ ] T012 (FR-004, FR-005, FR-006, FR-007, FR-003) [P] Follow quickstart smoke steps (`./ksail --help`, `./ksail cluster --help`, `./ksail cluster status`, and ensure `./ksail up` now errors) capturing any output adjustments for future release notes. *(Depends on: T011)*
- [ ] T013 (NFR-Lint) [P] Run `golangci-lint run` to ensure lint gates still pass after refactor. *(Depends on: T010)*
- [ ] T014 (FR-008) Ensure `reconcile` is not moved under `ksail cluster` and remains at the top level until migrated to `ksail workloads reconcile` in a future feature.

## Dependencies

- T002 → T003 → T004 → T005 (progressively building test coverage ahead of implementation)
- T005 → T006 → T007 ensures implementation follows the agreed test fixtures
- T007 → T008 handles removal of legacy wiring prior to documentation updates in T009
- T009 feeds into verification tasks (T010–T013) so help text and quickstart content are current before validation
- T010 must follow implementation before verification tasks T011–T013
- T011 must complete before non-functional validation tasks T015–T017

## Parallel Execution Example

```text
# After completing T011, these polish tasks can run together:
Task: T012 [P] Follow quickstart smoke steps (runtime validation)
Task: T013 [P] Run golangci-lint run
Task: T015 Validate CLI performance targets
Task: T016 Confirm golangci-lint baseline remains clean
Task: T017 Verify >90% coverage via go test reports
```

## Notes

- Keep tests failing until the corresponding implementation task completes (strict TDD).
- When moving files in T005, preserve git history with `git mv` where possible to ease review.
- Snapshot updates in T003 should accompany explanatory comments so reviewers understand expected help output changes.
- Document updates (T009) must mention the retirement of top-level lifecycle commands to avoid user confusion.
