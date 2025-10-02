# Tasks: CLI Command Timing

**Input**: Design documents from `/specs/005-implement-timing-in/`
**Prerequisites**: plan.md (required), research.md, data-model.md, contracts/

## Execution Flow (main)

```text
1. Load plan.md from feature directory
   → Tech stack: Go 1.24.0+, standard library time package, mockery
   → Structure: Single project (CLI tool with package-first design)
2. Load design documents:
   → data-model.md: Timer interface, TimingData struct
   → contracts/: timer-interface.md (7 requirements), notify-integration.md (5 requirements)
   → research.md: Design decisions, integration patterns
   → quickstart.md: 4 test scenarios for validation
3. Generate tasks by category:
   → Setup: Package structure, mockery config
   → Tests: 12 contract tests (7 timer + 5 notify integration)
   → Core: Timer implementation, FormatTiming helper
   → Integration: 8 CLI command updates
   → Polish: Documentation, performance validation
4. Apply task rules:
   → Contract tests = [P] (different test files)
   → Timer package = sequential (same files)
   → CLI commands = [P] (different command files)
5. Number tasks sequentially (T001-T025)
6. TDD ordering: ALL tests before implementation
```

## Format: `[ID] [P?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- Include exact file paths in descriptions

## Path Conventions

This is a single-project Go CLI tool with the following structure:

- `pkg/ui/timer/` - New timer package
- `pkg/ui/notify/` - Existing notify package (to be updated)
- `cmd/` - CLI commands (init, up, down, start, stop, list, status, reconcile)

## Phase 3.1: Setup

- [ ] **T001** Create package structure for `pkg/ui/timer/`
  - Create directory: `pkg/ui/timer/`
  - Create placeholder files: `timer.go`, `timer_test.go`, `doc.go`, `README.md`
  - Initialize with package documentation

- [ ] **T002** [P] Configure mockery for Timer interface
  - Update `.mockery.yml` to include `pkg/ui/timer` interfaces
  - Verify mockery can generate mocks for Timer interface

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

> **CRITICAL**: These tests MUST be written and MUST FAIL before ANY implementation

### Timer Interface Contract Tests (from timer-interface.md)

- [ ] **T003** [P] Contract test CR-001 (Start initialization) in `pkg/ui/timer/timer_test.go`
  - Test: Start() initializes both total and stage timers
  - Verify: GetTiming() returns durations ≈ 0 after Start()
  - Test: Multiple Start() calls reset the timer

- [ ] **T004** [P] Contract test CR-002 (GetTiming before Start) in `pkg/ui/timer/timer_test.go`
  - Test: GetTiming() before Start() returns (0, 0) or panics gracefully
  - Verify: Error handling is clear and predictable

- [ ] **T005** [P] Contract test CR-003 (NewStage stage transition) in `pkg/ui/timer/timer_test.go`
  - Test: NewStage() resets stage timer while preserving total time
  - Verify: Total continues accumulating, stage resets
  - Test: Multiple NewStage() calls work correctly

- [ ] **T006** [P] Contract test CR-004 (GetTiming returns current state) in `pkg/ui/timer/timer_test.go`
  - Test: GetTiming() can be called multiple times
  - Verify: Each call returns updated durations without side effects
  - Test: Timer state is not modified by GetTiming()

- [ ] **T007** [P] Contract test CR-005 (Single-stage command) in `pkg/ui/timer/timer_test.go`
  - Test: Without NewStage() calls, total == stage
  - Verify: Single-stage timing produces equal durations

- [ ] **T008** [P] Contract test CR-006 (Stop method) in `pkg/ui/timer/timer_test.go`
  - Test: Stop() can be called without errors
  - Verify: Stop() doesn't affect GetTiming() results
  - Test: Multiple Stop() calls are safe

- [ ] **T009** [P] Contract test CR-007 (Duration precision) in `pkg/ui/timer/timer_test.go`
  - Test: Sub-millisecond operations handled correctly
  - Verify: Duration.String() formatting matches specification
  - Test: Long-running operations (seconds/minutes) format correctly

### Notify Integration Contract Tests (from notify-integration.md)

- [ ] **T010** [P] Contract test IR-001 (Timer independence) in `pkg/ui/timer/integration_test.go`
  - Test: Verify pkg/ui/timer has no imports of pkg/ui/notify
  - Use: Static analysis or import inspection
  - Verify: Clean architecture maintained

- [ ] **T011** [P] Contract test IR-002 (Timing format consistency) in `pkg/ui/notify/notify_test.go`
  - Test: FormatTiming() produces "[X total|Y stage]" for multi-stage
  - Test: FormatTiming() produces "[X]" for single-stage
  - Verify: Uses Duration.String() formatting

- [ ] **T012** [P] Contract test IR-003 (Optional timing display) in `pkg/ui/notify/notify_test.go`
  - Test: Success() works without timing parameter
  - Test: Success() appends timing when provided
  - Test: Empty timing string handled correctly

- [ ] **T013** [P] Contract test IR-004 (Command integration pattern) in `pkg/ui/timer/integration_test.go`
  - Test: Complete integration flow (Start → NewStage → GetTiming → Success)
  - Verify: Timer and notify work together correctly
  - Test: Multi-stage and single-stage patterns

- [ ] **T014** [P] Contract test IR-005 (Error cases no timing) in `pkg/ui/timer/integration_test.go`
  - Test: Error paths don't display timing
  - Verify: Timer state ignored on failure
  - Test: No cleanup needed on error

## Phase 3.3: Core Implementation (ONLY after tests are failing)

### Timer Package Implementation

- [ ] **T015** Implement Timer interface definition in `pkg/ui/timer/timer.go`
  - Define Timer interface with Start(), NewStage(), GetTiming(), Stop()
  - Add comprehensive GoDoc comments
  - Define any internal types (e.g., timerImpl struct)

- [ ] **T016** Implement Timer constructor in `pkg/ui/timer/timer.go`
  - Create New() function returning Timer interface
  - Initialize internal state (startTime, stageStartTime, currentStage)
  - Ensure zero-value safety

- [ ] **T017** Implement Start() method in `pkg/ui/timer/timer.go`
  - Set startTime and stageStartTime to time.Now()
  - Handle multiple Start() calls (reset behavior)
  - Pass contract test CR-001

- [ ] **T018** Implement NewStage() method in `pkg/ui/timer/timer.go`
  - Reset stageStartTime to time.Now()
  - Update currentStage with title
  - Preserve total elapsed time
  - Pass contract test CR-003

- [ ] **T019** Implement GetTiming() method in `pkg/ui/timer/timer.go`
  - Calculate total = time.Since(startTime)
  - Calculate stage = time.Since(stageStartTime)
  - Return both durations
  - Pass contract tests CR-004, CR-005, CR-007

- [ ] **T020** Implement Stop() method in `pkg/ui/timer/timer.go`
  - Optional cleanup for future extensibility
  - No-op implementation for now
  - Pass contract test CR-006

- [ ] **T021** Add package documentation in `pkg/ui/timer/doc.go` and `pkg/ui/timer/README.md`
  - Package-level GoDoc with examples
  - README with usage instructions
  - Code examples for single-stage and multi-stage usage
  - Follow constitutional package-first requirements

### Notify Package Updates

- [ ] **T022** Implement FormatTiming() helper in `pkg/ui/notify/notify.go`
  - Add FormatTiming(total, stage time.Duration, isMultiStage bool) string
  - Implement format logic: "[X]" or "[X total|Y stage]"
  - Add GoDoc comments
  - Pass contract test IR-002

- [ ] **T023** Update Success() function in `pkg/ui/notify/notify.go`
  - Add optional timing parameter: Success(message string, timing ...string)
  - Append timing to message if provided
  - Maintain backward compatibility
  - Pass contract test IR-003

## Phase 3.4: CLI Command Integration

**Note**: Each command update is parallelizable as they modify different files

- [ ] **T024** [P] Add timing to `init` command in `cmd/init.go`
  - Create timer, call Start()
  - Get timing on success: total, stage := timer.GetTiming()
  - Format timing: timingStr := notify.FormatTiming(total, stage, false)
  - Update success call: notify.Success(fmt.Sprintf("message %s", timingStr))

- [ ] **T025** [P] Add timing to `up` command in `cmd/up.go`
  - Create timer, call Start()
  - Add NewStage() calls at each operation phase
  - Get timing on success: total, stage := timer.GetTiming()
  - Format timing: timingStr := notify.FormatTiming(total, stage, true)
  - Update success call with timing

- [ ] **T026** [P] Add timing to `down` command in `cmd/down.go`
  - Create timer, call Start()
  - Add NewStage() calls for multi-stage teardown
  - Get timing and format for success message
  - Ensure no timing on error paths

- [ ] **T027** [P] Add timing to `start` command in `cmd/start.go`
  - Create timer, call Start()
  - Add NewStage() calls if applicable
  - Get timing and format for success message

- [ ] **T028** [P] Add timing to `stop` command in `cmd/stop.go`
  - Create timer, call Start()
  - Add NewStage() calls if applicable
  - Get timing and format for success message

- [ ] **T029** [P] Add timing to `list` command in `cmd/list.go`
  - Create timer, call Start()
  - Single-stage command (no NewStage calls)
  - Get timing and format for success message

- [ ] **T030** [P] Add timing to `status` command in `cmd/status.go`
  - Create timer, call Start()
  - Single-stage command (no NewStage calls)
  - Get timing and format for success message

- [ ] **T031** [P] Add timing to `reconcile` command in `cmd/reconcile.go`
  - Create timer, call Start()
  - Add NewStage() calls for reconciliation phases
  - Get timing and format for success message

## Phase 3.5: Polish & Validation

- [ ] **T032** [P] Run mockery to generate Timer interface mocks
  - Execute: `mockery`
  - Verify: Mock files generated in `pkg/ui/timer/mocks/`
  - Commit: Generated mocks

- [ ] **T033** [P] Performance validation for timer overhead
  - Measure timer Start() + GetTiming() overhead
  - Verify: <1ms overhead as per constitution requirement
  - Document: Performance characteristics in README.md

- [ ] **T034** [P] Execute Scenario 1 from `quickstart.md` (multi-stage command)
  - Run: `./ksail up`
  - Verify: Timing displayed after each stage
  - Verify: Format matches "[X total|Y stage]"
  - Document: Any deviations from expected output

- [ ] **T035** [P] Execute Scenario 2 from `quickstart.md` (single-stage command)
  - Run: `./ksail init --distribution Kind`
  - Verify: Timing displayed in format "[X]"
  - Verify: Sub-second precision visible

- [ ] **T036** [P] Execute Scenario 3 from `quickstart.md` (command failure)
  - Run: `./ksail init --distribution InvalidDistribution`
  - Verify: NO timing information in error output
  - Verify: Non-zero exit code

- [ ] **T037** [P] Execute Scenario 4 from `quickstart.md` (long-running command)
  - Run: `./ksail up` then `./ksail down`
  - Verify: Timing uses appropriate units (seconds/minutes)
  - Verify: Progressive timing updates

- [ ] **T038** Run linter and fix any issues
  - Execute: `golangci-lint run`
  - Fix: Any linting errors in timer package or CLI commands
  - Verify: Clean lint output

- [ ] **T039** Run full test suite
  - Execute: `go test ./...`
  - Verify: All tests pass (contract tests + integration tests)
  - Fix: Any test failures

- [ ] **T040** Update `.github/copilot-instructions.md` with timing feature
  - Document: Timer package usage patterns
  - Document: Integration examples with CLI commands
  - Ensure: Constitutional compliance noted

## Dependencies

```text
Setup (T001-T002)
  ↓
Tests First (T003-T014) - ALL must be written and failing
  ↓
Core Implementation (T015-T023)
  ↓
CLI Integration (T024-T031) - can run in parallel
  ↓
Polish & Validation (T032-T040) - mostly parallel
```

**Blocking Relationships**:

- T003-T014 block T015-T023 (TDD: tests before implementation)
- T015-T023 block T024-T031 (need timer package before CLI integration)
- T024-T031 block T034-T037 (need CLI updates before quickstart validation)
- T032 (mockery) can run after T015 (interface defined)

## Parallel Execution Examples

### Phase 3.2: Contract Tests (All Parallel)

```bash
# All timer contract tests (T003-T009) can run together
go test -run TestCR001 ./pkg/ui/timer/  # T003
go test -run TestCR002 ./pkg/ui/timer/  # T004
go test -run TestCR003 ./pkg/ui/timer/  # T005
# ... etc

# All notify integration tests (T010-T014) can run together
go test -run TestIR001 ./pkg/ui/timer/  # T010
go test -run TestIR002 ./pkg/ui/notify/ # T011
# ... etc
```

### Phase 3.4: CLI Integration (All Parallel)

```bash
# All CLI command updates (T024-T031) are independent
# Can be assigned to different developers or AI agents
# Task: "Add timing to init command in cmd/init.go"
# Task: "Add timing to up command in cmd/up.go"
# Task: "Add timing to down command in cmd/down.go"
# ... etc
```

### Phase 3.5: Validation (Mostly Parallel)

```bash
# Quickstart scenarios (T034-T037) can run in sequence or parallel
# Performance test (T033) can run independently
# Mockery (T032) can run after T015
```

## Notes

- **[P] tasks**: Different files, no dependencies, can run in parallel
- **TDD Mandatory**: ALL tests (T003-T014) must fail before implementing (T015-T023)
- **Constitutional Compliance**: Timer package follows package-first design, <1ms overhead
- **Integration Pattern**: Timer → GetTiming() → notify.FormatTiming() → notify.Success()
- **Error Handling**: No timing displayed on command failures (per IR-005)

## Task Generation Rules Applied

1. **From Contracts**:
   - 7 timer contract tests (CR-001 through CR-007) → T003-T009
   - 5 notify integration tests (IR-001 through IR-005) → T010-T014

2. **From Data Model**:
   - Timer entity → T015-T020 (interface + implementation)
   - TimingData (implicit in GetTiming return) → handled in T019

3. **From Quickstart**:
   - 4 test scenarios → T034-T037 validation tasks

4. **From Research**:
   - Timer design pattern → T015-T020
   - Notify integration → T022-T023
   - CLI integration pattern → T024-T031

5. **Ordering**:
   - Setup (T001-T002) → Tests (T003-T014) → Implementation (T015-T023) → Integration (T024-T031) → Polish (T032-T040)

## Validation Checklist

- [x] All contracts have corresponding tests (12 tests for 12 contract requirements)
- [x] All entities have implementation tasks (Timer interface + impl)
- [x] All tests come before implementation (T003-T014 before T015-T023)
- [x] Parallel tasks truly independent (different files marked [P])
- [x] Each task specifies exact file path
- [x] No task modifies same file as another [P] task
- [x] TDD workflow enforced (tests must fail before implementation)
- [x] Constitutional requirements addressed (package-first, <1ms overhead)

---

**Total Tasks**: 40 (2 setup + 12 tests + 7 implementation + 8 CLI integration + 11 polish)
**Estimated Duration**: 3-5 days (depends on parallel execution and team size)
**Critical Path**: T001 → T003-T014 (tests) → T015-T020 (timer impl) → T024-T031 (CLI) → T039 (validation)
