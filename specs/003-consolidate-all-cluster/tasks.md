# Tasks: Consolidate Cluster Commands Under `ksail cluster`

**Note:** Do not move `reconcile` under `ksail cluster` in this refactor. Leave it at the top level until it is migrated to `ksail workloads reconcile` in a future change.

**Input**: Design documents from `/specs/003-consolidate-all-cluster/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Requirement Coverage Checklist

| Requirement Key | Description                                                           | Covered By Task(s)  |
| --------------- | --------------------------------------------------------------------- | ------------------- |
| FR-001          | Expose parent `cluster` command                                       | T001, T005, T006    |
| FR-002          | Subcommands for all cluster lifecycle actions (excluding 'reconcile') | T005, T006, T011    |
| FR-003          | Subcommand behavior/messaging parity                                  | T011, T012          |
| FR-004          | `ksail cluster --help` guidance                                       | T012, T003          |
| FR-005          | `ksail cluster` (bare) help output                                    | Explicit test, T012 |
| FR-006          | Remove legacy top-level commands                                      | T008, T012          |
| FR-007          | Root help lists `cluster`                                             | T007, T012          |
| FR-008          | Exclude `reconcile` from cluster refactor                             | T014                |

## Non-Functional Validation

- [ ] T015 (FR-Performance) Validate CLI response and performance meet defined thresholds
- [ ] T016 (FR-Lint) Validate all code and tests pass golangci-lint with zero issues
- [ ] T017 (FR-Coverage) Validate >90% test coverage is achieved and validated

**Note:** The `reconcile` command is intentionally excluded from this refactor. It will be moved to `ksail workloads reconcile` in a future change.

## Phase 3.1: Setup

- [ ] T001 (FR-001) Prepare `cmd/cluster/` package scaffold (create directory and lightweight `doc.go` explaining the consolidated cluster command namespace) so subsequent tests can target the new package.

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

- [ ] T002 (FR-005, FR-003) Add explicit test for `ksail cluster` (bare) help output and ensure messaging parity with legacy commands (TDD)

## Phase 3.4: Integration & Validation

- [ ] T010 (FR-Lint) [P] Run `gofmt`/`goimports` on all touched Go files (including the new `cmd/cluster/` directory) to satisfy formatting gates. *(Depends on: T007)*
- [ ] T011 (FR-002, FR-003, FR-007, FR-Performance, FR-Coverage) Execute `go test ./cmd` (updating go-snaps snapshots as needed) followed by `go test ./...` to confirm all suites pass with the new command structure. *(Depends on: T010)*
- [ ] T012 (FR-004, FR-005, FR-006, FR-007, FR-003) [P] Follow quickstart smoke steps (`./ksail --help`, `./ksail cluster --help`, `./ksail cluster status`, and ensure `./ksail up` now errors) capturing any output adjustments for future release notes. *(Depends on: T011)*
- [ ] T013 (FR-Lint) [P] Run `golangci-lint run` to ensure lint gates still pass after refactor. *(Depends on: T010)*
- [ ] T014 (FR-008) Ensure `reconcile` is not moved under `ksail cluster` and remains at the top level until migrated to `ksail workloads reconcile` in a future feature.

## Dependencies

- T002 → T003 (snapshot expectations depend on updated tests)
- T004 → T005 (tests define expected structure before moving code)
- T005 → T006 → T007 cascades the command consolidation implementation
- T007 → T008 → T009 handles documentation cleanup
- T010 must follow implementation before verification tasks T011–T013

## Parallel Execution Example

```text
# After completing T011, these polish tasks can run together:
Task: T012 [P] Follow quickstart smoke steps (runtime validation)
Task: T013 [P] Run golangci-lint run
```

## Notes

- Keep tests failing until the corresponding implementation task completes (strict TDD).
- When moving files in T005, preserve git history with `git mv` where possible to ease review.
- Snapshot updates in T003 should accompany explanatory comments so reviewers understand expected help output changes.
- Document updates (T009) must mention the retirement of top-level lifecycle commands to avoid user confusion.
