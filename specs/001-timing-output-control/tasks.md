# Tasks: Timing Output Control

**Input**: Design documents from `/specs/001-timing-output-control/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/, quickstart.md

**Tests**: REQUIRED (constitution: test-first). Only omit tests for docs-only changes.

**No white-box tests**: Tests MUST validate exported behavior only. Do not test unexported functions/fields or internal implementation details.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Establish a baseline and identify impacted CLI outputs.

- [ ] T001 Run baseline quality gates in `projects/ksail-go/` (`go test ./...`, `go build ./...`, `golangci-lint run --timeout 5m`)
- [ ] T002 [P] Inventory existing timing output call sites in `projects/ksail-go/` (grep for `notify.FormatTiming(` and `Message{Timer: ...}`)
- [ ] T003 [P] Inventory snapshot expectations containing legacy timing suffix in `projects/ksail-go/cmd/__snapshots__/root_test.snap`, `projects/ksail-go/cmd/__snapshots__/init_test.snap`, `projects/ksail-go/cmd/cluster/__snapshots__/init_test.snap`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Add shared plumbing and formatting helpers used by all commands.

**⚠️ CRITICAL**: No user story work can begin until this phase is complete.

- [ ] T004 Add persistent `--timing` flag to root command in `projects/ksail-go/cmd/root.go`
- [ ] T005 [P] Add public helper to read timing flag state (and/or context) in `projects/ksail-go/pkg/cmd/timing.go`
- [ ] T006 [P] Add timing block formatter(s) for the spec format in `projects/ksail-go/pkg/ui/notify/notify.go` (e.g., `FormatTimingBlock(current, total time.Duration) string`)
- [ ] T007 [P] Add notify helper to print the timing block after a completion message in `projects/ksail-go/pkg/ui/notify/notify.go` (writer-aware; no direct stdout)
- [ ] T008 [P] Add public-API tests for the new notify formatter(s) in `projects/ksail-go/pkg/ui/notify/notify_test.go`

**Checkpoint**: Foundation ready — commands can now be updated story-by-story.

---

## Phase 3: User Story 1 — Enable timing for a single run (Priority: P1) 🎯 MVP

**Goal**: Users can opt into timing output with `--timing` (default off).

**Independent Test**: Run the same command with and without `--timing`; only the `--timing` run prints the timing block.

### Tests for User Story 1 (REQUIRED) ⚠️

> Write these tests FIRST and ensure they FAIL before implementing behavior.

- [ ] T009 [P] [US1] Update root help snapshot to include `--timing` flag in `projects/ksail-go/cmd/__snapshots__/root_test.snap`
- [ ] T010 [P] [US1] Add/extend root command snapshot coverage for `--timing` presence in `projects/ksail-go/cmd/root_test.go`
- [ ] T011 [P] [US1] Update init command snapshots to remove default timing suffix in `projects/ksail-go/cmd/__snapshots__/init_test.snap` and `projects/ksail-go/cmd/cluster/__snapshots__/init_test.snap`
- [ ] T012 [P] [US1] Add new snapshot test case for `--timing` on init to validate the 3-line timing block format in `projects/ksail-go/cmd/init_test.go`
- [ ] T013 [P] [US1] Add new snapshot test case for `--timing` on cluster init to validate the 3-line timing block format in `projects/ksail-go/cmd/cluster/init_test.go`

### Implementation for User Story 1

- [ ] T014 [US1] Wire `--timing` flag value into command execution (e.g., via context) in `projects/ksail-go/cmd/root.go`
- [ ] T015 [US1] Stop appending legacy timing suffix to init completion output unless `--timing` is enabled in `projects/ksail-go/cmd/init.go`
- [ ] T016 [US1] When `--timing` is enabled, print the spec timing block after the init completion message in `projects/ksail-go/cmd/init.go`
- [ ] T017 [US1] Ensure stage boundaries are correct for init (use `Timer.Start()` and `Timer.NewStage()` appropriately) in `projects/ksail-go/cmd/init.go`

**Checkpoint**: `ksail --timing init` (and `ksail init`) behaves per spec and is independently testable.

---

## Phase 4: User Story 2 — Consistent timing output format (Priority: P3)

**Goal**: Any command that emits `✔` completion messages uses the same timing block format and prints it after each timed activity.

**Independent Test**: Enable timing for a multi-stage command and verify each completion message is followed by a timing block; `current` reflects the most recent stage, `total` accumulates.

### Tests for User Story 2 (REQUIRED) ⚠️

- [ ] T018 [P] [US2] Update notify timing tests away from legacy `"[stage: ...]"` usage where needed in `projects/ksail-go/pkg/ui/notify/notify_test.go`
- [ ] T019 [P] [US2] Add snapshot coverage for a multi-stage command with `--timing` (pick an existing test harness) in `projects/ksail-go/cmd/cluster/create_test.go` (or the nearest existing create tests)

### Implementation for User Story 2

- [ ] T020 [US2] Refactor lifecycle success output to stop concatenating `notify.FormatTiming(...)` into the same line in `projects/ksail-go/pkg/cmd/lifecycle_helpers.go`
- [ ] T021 [US2] Update lifecycle flow to print timing block only when `--timing` is enabled, after each `✔` completion message, in `projects/ksail-go/pkg/cmd/lifecycle_helpers.go`
- [ ] T022 [US2] Refactor cluster create to remove `notify.FormatTiming(...)` suffixes and use the timing block when enabled in `projects/ksail-go/cmd/cluster/create.go`
- [ ] T023 [US2] Refactor workload reconcile to remove `notify.FormatTiming(...)` suffix usage and use the timing block when enabled in `projects/ksail-go/cmd/workload/reconcile.go`
- [ ] T024 [US2] Refactor workload helm release generation to remove `notify.FormatTiming(...)` suffix usage and use the timing block when enabled in `projects/ksail-go/cmd/workload/gen/helm_release.go`
- [ ] T025 [US2] Ensure each refactored command creates stage boundaries per timed activity (call `Timer.NewStage()` after emitting completion+timing) in `projects/ksail-go/cmd/cluster/create.go`, `projects/ksail-go/cmd/workload/reconcile.go`, `projects/ksail-go/cmd/workload/gen/helm_release.go`, `projects/ksail-go/pkg/cmd/lifecycle_helpers.go`

**Checkpoint**: With `--timing`, timing blocks appear after each `✔ ...` across the updated multi-stage flows.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Consistency, docs, and final validation.

- [ ] T026 [P] Update user-facing documentation to mention `--timing` and show the output format in `projects/ksail-go/README.md`
- [ ] T027 [P] Update quickstart instructions for this feature in `projects/ksail-go/specs/001-timing-output-control/quickstart.md` if implementation details changed
- [ ] T028 Regenerate/refresh snapshots after output changes in `projects/ksail-go/cmd/__snapshots__/root_test.snap`, `projects/ksail-go/cmd/__snapshots__/init_test.snap`, `projects/ksail-go/cmd/cluster/__snapshots__/init_test.snap`
- [ ] T029 Run full quality gates after implementation in `projects/ksail-go/` (`mockery`, `go test ./...`, `golangci-lint run --timeout 5m`, `go build ./...`)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies
- **Foundational (Phase 2)**: Depends on Setup; BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational
- **User Story 2 (Phase 4)**: Depends on US1 (so MVP is proven first)
- **Polish (Phase 5)**: Depends on desired story completion

### User Story Dependencies

- **US1 (P1)**: Can start after Foundational — produces the MVP for `--timing`.
- **US2 (P3)**: Extends coverage across additional commands and removes legacy suffix formatting.

---

## Parallel Opportunities

- Phase 2 tasks T005–T008 can be parallelized (different files).
- In Phase 3, snapshot edits (T009–T013) can run in parallel.
- In Phase 4, the refactors (T020–T024) can be split across team members (different command files).

---

## Parallel Example: User Story 1

```bash
Task: "Update root help snapshot in cmd/__snapshots__/root_test.snap"
Task: "Update init snapshots in cmd/__snapshots__/init_test.snap and cmd/cluster/__snapshots__/init_test.snap"
Task: "Add init snapshot tests in cmd/init_test.go and cmd/cluster/init_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1–2
2. Complete Phase 3 (US1)
3. Validate: `ksail init` (no timing) and `ksail --timing init` (timing block)

### Incremental Delivery

1. US1 proves flag + formatting + one command path.
2. US2 rolls out consistent formatting and removes legacy suffixes across additional commands.
3. Polish updates docs and runs all gates.
