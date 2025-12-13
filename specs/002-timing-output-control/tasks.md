---
description: "Task list for Timing Output Control"
---

# Tasks: Timing Output Control

**Input**: Design documents from `/specs/002-timing-output-control/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Required (constitution + FR-010). Tests MUST validate exported/public behavior only.

**No white-box tests**: Do not test unexported functions/fields or internal implementation details.

## Format: `- [ ] T### [P?] [US#?] Description with file path`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[US#]**: User story label (required for story phases only)

---

## Phase 1: Setup (Shared Infrastructure)

- [x] T001 Confirm baseline passes: run `go test ./...` from repo root (go.mod)
- [x] T002 Confirm baseline passes: run `go build ./...` from repo root (go.mod)
- [x] T003 Run `mockery` to ensure mocks generate cleanly before changes (go.mod)
- [x] T004 Inventory timing-related output callsites to change (pkg/ui/notify/notify.go, cmd/root.go, cmd/cluster/create.go, cmd/cluster/delete.go, cmd/cluster/init.go, cmd/workload/reconcile.go, pkg/cmd/lifecycle_helpers.go, pkg/io/config-manager/ksail/manager.go)

---

## Phase 2: Foundational (Blocking Prerequisites)

- [x] T005 Add global/root persistent `--timing` flag (default false) in cmd/root.go
- [x] T006 [P] Add helper to read the timing flag consistently (pkg/cmd/flags.go)
- [x] T007 [P] Document the flag in repo docs (README.md)

**Checkpoint**: `--timing` exists globally and is discoverable in docs.

---

## Phase 3: User Story 1 - Default Output (Priority: P1) 🎯 MVP

**Goal**: Timing output stays hidden by default; existing output remains stable.

**Independent Test**: Run representative commands without `--timing` and assert no `⏲` timing lines are printed.

### Tests (write first; ensure failing before implementation)

- [x] T008 [P] [US1] Add CLI behavior/snapshot test asserting default output contains no `⏲` (cmd/root_test.go)
- [x] T009 [P] [US1] Add unit test asserting `--timing` default is false (cmd/root_test.go)

### Implementation

- [x] T010 [US1] Gate timing output off by default by not attaching timers unless `--timing` is set (cmd/cluster/create.go, cmd/cluster/delete.go, cmd/cluster/init.go, cmd/workload/reconcile.go)
- [x] T011 [US1] Update non-cmd code paths that emit success messages with timers to respect `--timing` (pkg/io/config-manager/ksail/manager.go, pkg/cmd/lifecycle_helpers.go)


**Checkpoint**: Running the same commands without `--timing` produces unchanged output (no timing lines).

---

## Phase 4: User Story 2 - On-Demand Timing (Priority: P2)

**Goal**: When `--timing` is enabled, emit a timing block after each completed activity/stage success message.

**Independent Test**: Run a multi-stage command with `--timing` enabled and assert per-activity timing blocks with monotonic totals.

### Tests (write first; ensure failing before implementation)

- [x] T012 [P] [US2] Add notify rendering test for required timing block format with `⏲ current` + `total` lines (pkg/ui/notify/notify_test.go)
- [x] T013 [P] [US2] Add notify rendering test ensuring the `✔` line remains unchanged and timing block prints immediately after it (pkg/ui/notify/notify_test.go)
- [x] T014 [P] [US2] Add CLI test asserting `--timing` enables timing output lines (cmd/root_test.go)
- [x] T015 [P] [US2] Add error-path test ensuring timing output is not printed on failure (cmd/root_test.go)

### Implementation

- [x] T016 [US2] Replace inline timing suffix formatting with a multi-line timing block renderer (pkg/ui/notify/notify.go)
- [x] T017 [US2] Ensure durations use Go `time.Duration` string formatting in the timing renderer (pkg/ui/notify/notify.go)
- [x] T018 [US2] Ensure timing is emitted after each completed activity/stage success message when enabled (cmd/cluster/create.go, cmd/cluster/delete.go, cmd/cluster/init.go, cmd/workload/reconcile.go, pkg/cmd/lifecycle_helpers.go)
- [x] T019 [US2] Ensure `total` is monotonically non-decreasing across multi-stage runs when enabled (pkg/ui/timer/timer.go, pkg/ui/notify/notify.go)
- [x] T020 [US2] Ensure timing output is not printed on errors (pkg/ui/notify/notify.go, cmd/cluster/create.go, cmd/cluster/delete.go, cmd/cluster/init.go, cmd/workload/reconcile.go, pkg/cmd/lifecycle_helpers.go)

**Checkpoint**: With `--timing`, each stage completion prints the timing block; without `--timing`, no timing lines appear.

---

## Phase 5: Polish & Cross-Cutting Concerns

- [x] T021 [P] Update any user-facing docs that mention timing output format or behavior (README.md, docs/)
- [x] T022 Run formatter: `golangci-lint fmt` (go.mod)
- [x] T023 Run linter: `golangci-lint run --timeout 5m` (go.mod)
- [x] T024 Run full test suite: `go test ./...` (go.mod)
- [x] T025 Run full build: `go build ./...` (go.mod)

---

## Dependencies & Execution Order

### User Story Dependencies

- **US1 (P1)**: Depends on Phase 2 (flag exists) and then ensures default output stays unchanged.
- **US2 (P2)**: Depends on Phase 2 and can reuse US1’s gating behavior; adds new rendering + tests.

### Within Each User Story

- Tests MUST be written and FAIL before implementation.
- Rendering/unit tests (`pkg/ui/notify`) can be developed in parallel with CLI behavior tests (`cmd/*_test.go`).

---

## Parallel Opportunities

- [P] tasks can be done concurrently (different files/no direct dependency).
- In US2, notify tests and CLI tests are parallel once the acceptance criteria are clear.

---

## Parallel Example: User Story 1

```text
In parallel:
- T008 (default output snapshot test) in cmd/root_test.go
- T009 (default flag value test) in cmd/root_test.go
```

---

## Parallel Example: User Story 2

```text
In parallel:
- T012 (notify format test) in pkg/ui/notify/notify_test.go
- T014 (CLI flag behavior test) in cmd/root_test.go
- T015 (error behavior test) in cmd/root_test.go
```

---

## Implementation Strategy

### MVP First (US1 Only)

1. Phase 1 → Phase 2
2. Phase 3 (US1)
3. Validate: run the CLI without `--timing` and confirm no timing output

### Incremental Delivery

- Add US2 after US1 is stable: implement the new renderer + wiring and re-validate output stability when timing is off.
