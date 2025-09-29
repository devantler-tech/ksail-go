# Tasks: KSail Cluster Provisioning Command

**Input**: Design documents from `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Phase 3.1: Setup

- [ ] T001 Run `/Users/ndam/git-personal/monorepo/projects/ksail-go/scripts/run-mockery.sh` to refresh provisioner mocks before adding new tests.

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

> CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation.

- [ ] T002 Create CLI contract tests covering success output and dependency failure messaging (including telemetry summary fields) in `cmd/cluster/up_contract_test.go`.
- [ ] T003 Add Kind, K3d, EKS, and `--force` quickstart integration scenarios to `cmd/cluster/up_integration_test.go`, stubbing provisioners and asserting kubeconfig/context handling.
- [ ] T004 Extend `cmd/cluster/up_internal_test.go` with unit tests for configuration precedence resolution, telemetry stage transitions, and readiness/kubeconfig failure paths.

## Phase 3.3: Core Implementation (ONLY after tests are failing)

- [ ] T005 Implement a configuration precedence loader in `cmd/cluster/up.go` that materialises the Combined Configuration Snapshot (flags → env → files → defaults).
- [ ] T006 Wire distribution-specific config managers and provisioners in `cmd/cluster/up.go`, reusing existing packages and enforcing dependency checks before provisioning.
- [ ] T007 Implement `--force` recreation semantics and idempotent reuse handling in `cmd/cluster/up.go`, coordinating `Exists`, `Delete`, `Create`, and `Start` flows.
- [ ] T008 Integrate readiness polling and kubeconfig merge logic in `cmd/cluster/up.go`, surfacing actionable timeout/merge errors.
- [ ] T009 Emit telemetry summaries and structured success/failure notifications in `cmd/cluster/up.go`, including slowest-stage reporting.

## Phase 3.4: Integration & Verification

- [ ] T010 Finalise Cobra flag bindings and Viper configuration wiring in `cmd/cluster/up.go` and related helpers to ensure effective values propagate to provisioning.

## Phase 3.5: Polish

- [ ] T011 Run gofmt on `cmd/cluster/up.go` and execute `go test ./cmd/cluster` to confirm green build.
- [ ] T012 [P] Update telemetry/output documentation in `README.md` (and related docs) to explain configuration precedence and stage summaries.
- [ ] T013 [P] Follow `quickstart.md` scenarios manually (Kind, K3d, EKS, `--force`) capturing timing summaries for release validation.

## Dependencies

- T002–T004 must complete (and initially fail) before starting T005–T010.
- T005 enables T006, which must finish before T007.
- T007 must complete before T008; T008 must complete before T009.
- T010 depends on the successful completion of T005–T009.
- Polish tasks (T011–T013) begin after all implementation tasks succeed.

## Parallel Execution Example

```bash
# After implementation succeeds, run documentation and manual validation in parallel:
Task: "T012 [P] Update telemetry/output documentation in README.md to explain configuration precedence and stage summaries."
Task: "T013 [P] Follow quickstart.md scenarios manually (Kind, K3d, EKS, --force) capturing timing summaries for release validation."
```

## Validation Checklist

- [x] All contracts have corresponding test tasks (T002).
- [x] Entities and telemetry/data structures covered by test tasks (T004) and implementation tasks (T005–T009).
- [x] Tests precede implementation tasks in ordering.
- [x] Parallel tasks (T012, T013) operate on distinct artefacts.
- [x] Each task specifies explicit file paths or scripts.
- [x] No [P] tasks modify the same file.
- [x] Telemetry work is captured in T002, T004, and T009.
