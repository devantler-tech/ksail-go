# Tasks: KSail Cluster Provisioning Command

**Input**: Design documents from `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/005-implement-the-description/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/, quickstart.md

## Phase 3.1: Setup

- [x] T001 Run `/Users/ndam/git-personal/monorepo/projects/ksail-go/scripts/run-mockery.sh` to refresh provisioner mocks before adding new tests.

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

> CRITICAL: These tests MUST be written and MUST FAIL before ANY implementation.

- [x] T002 Create CLI contract tests in `cmd/cluster/up_test.go` covering success output, dependency failures, validation errors, timeout behavior, and telemetry reporting per `contracts/cluster-up.md`.

## Phase 3.3: Core Implementation (ONLY after tests are failing)

- [x] T005 Implement configuration precedence and validation in `cmd/cluster/up.go` that materializes the Combined Configuration Snapshot (flags → env → files → defaults).
- [x] T006 Wire distribution-specific config managers and provisioners in `cmd/cluster/up.go`, reusing existing packages and enforcing dependency checks before provisioning.
- [x] T007 Implement `--force` recreation semantics and idempotent reuse handling in `cmd/cluster/up.go`, coordinating `Exists`, `Delete`, `Create`, and `Start` flows.
- [ ] T008 Integrate readiness polling and kubeconfig merge logic in `cmd/cluster/up.go`, surfacing actionable timeout/merge errors.
- [x] T009 Emit telemetry summaries and structured success/failure notifications in `cmd/cluster/up.go`, including slowest-stage reporting.

## Phase 3.4: Integration & Verification

- [ ] T010 Finalise Cobra flag bindings and Viper configuration wiring in `cmd/cluster/up.go` and related helpers to ensure effective values propagate to provisioning.

## Phase 3.5: Polish

- [ ] T011 Run gofmt on `cmd/cluster/up.go` and execute `go test ./cmd/cluster` to confirm green build.
- [ ] T012 [P] Update telemetry/output documentation in `README.md` to explain configuration precedence and stage summaries.
- [ ] T013 [P] Follow `quickstart.md` scenarios manually (Kind, K3d, EKS, `--force`) capturing timing summaries for release validation.

## Dependencies

- T002 must complete (and initially fail) before starting T005–T010.
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
- [x] Contract tests cover success output, dependency failures, validation errors, timeout behavior, and telemetry per `contracts/cluster-up.md`.
- [x] Tests precede implementation tasks in ordering.
- [x] Parallel tasks (T012, T013) operate on distinct artefacts.
- [x] Each task specifies explicit file paths or scripts.
- [x] No [P] tasks modify the same file.
- [x] Telemetry work is captured in T002 and T009.
