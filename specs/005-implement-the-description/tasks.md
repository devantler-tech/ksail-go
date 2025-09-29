# Tasks: KSail Cluster Provisioning Command

**Input**: Design documents from `/specs/005-implement-the-description/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/cluster-up.md, quickstart.md

## Phase 3.1: Setup

- [ ] T001 Introduce helper scaffolding in `cmd/cluster/up.go` (type stubs for dependency results, telemetry summary, readiness config, TODO markers removed).
  - Dependencies: Implementation plan approved (`plan.md`)
  - Notes: Keep logic within the command package; add doc comments outlining helper responsibilities.
- [ ] T002 [P] Prepare shared test fixtures under `cmd/cluster/testdata/` (kubeconfig templates, fake node payloads, AWS profile stubs).
  - Dependencies: T001
  - Notes: Provide reusable files referenced by dependency, readiness, telemetry, and kubeconfig tests.

## Phase 3.2: Tests First (TDD)

- [ ] T003 [P] Author failing dependency-check tests in `cmd/cluster/up_internal_test.go` covering Docker/Podman detection and AWS credential errors.
  - Dependencies: T001, T002
- [ ] T004 [P] Author failing readiness waiter tests in `cmd/cluster/up_internal_test.go` for success, timeout, and API failure flows using fake clientsets.
  - Dependencies: T001, T002
- [ ] T005 [P] Author failing kubeconfig management tests in `cmd/cluster/up_internal_test.go` validating context creation, switching, and persistence to disk.
  - Dependencies: T001, T002
- [ ] T006 Author failing orchestration helper tests in `cmd/cluster/up_internal_test.go` covering new cluster creation, reuse, force recreate, dependency failure, readiness timeout, and telemetry integration.
  - Dependencies: T003, T004, T005
- [ ] T007 [P] Expand `cmd/cluster/up_test.go` coverage for CLI flag wiring, helper invocation, dependency error messaging, readiness timeout exit codes, and confirmation that success output routes through existing notify/provisioner helpers with telemetry summary.
  - Dependencies: T006
- [ ] T008 [P] Add timing instrumentation tests in `cmd/cluster/up_internal_test.go` to validate per-stage duration capture, slowest-stage computation, and summary formatting (including quiet/JSON modes stubs).
  - Dependencies: T003, T004, T005

## Phase 3.3: Core Implementation (after tests are failing)

- [ ] T009 Implement config loading helpers in `cmd/cluster/up.go` returning validated `v1alpha1.Cluster` and provider configs expected by tests.
  - Dependencies: T006
- [ ] T010 Implement dependency checking helper and `dependencyCheckResult` struct in `cmd/cluster/up.go` to satisfy T003 expectations.
  - Dependencies: T003
- [ ] T011 Implement readiness waiter helper with `readinessProbeConfig` in `cmd/cluster/up.go` using client-go polling to satisfy T004.
  - Dependencies: T004
- [ ] T012 Implement kubeconfig management helper in `cmd/cluster/up.go` ensuring context creation and persistence to satisfy T005.
  - Dependencies: T005
- [ ] T013 Implement telemetry helper and orchestration flow in `cmd/cluster/up.go`, wiring provisioning logic, force semantics, and telemetry summary to satisfy T006/T008.
  - Dependencies: T009, T010, T011, T012
- [ ] T014 Finalize CLI handler in `cmd/cluster/up.go` to parse flags, call helpers, and surface notify/provisioner output expected by T007.
  - Dependencies: T013, T007

## Phase 3.4: Integration

- [ ] T015 Expose helper seams (function parameters/interfaces) inside `cmd/cluster/up.go` so tests can inject fakes while the CLI uses real provisioners and config managers.
  - Dependencies: T013, T007
- [ ] T016 Validate real provisioner wiring in `cmd/cluster/up.go`, ensuring Kind/K3d/EKS paths invoke existing packages and respect `--force` semantics.
  - Dependencies: T013, T015

## Phase 3.5: Polish

- [ ] T017 [P] Update user-facing documentation (`README.md` and related sections) to reflect new `ksail cluster up` behaviour and quickstart guidance.
  - Dependencies: T014, T016
- [ ] T018 [P] Run `go test ./... -coverpkg=./... -coverprofile=coverage.out`, refresh generated artifacts (mocks/snaps), and fail the task if overall coverage drops below 90% (capture `go tool cover -func=coverage.out`).
  - Dependencies: T009, T010, T011, T012, T013, T014, T015, T016
- [ ] T019 [P] Run `golangci-lint run --timeout 5m` and resolve lint issues introduced by the feature.
  - Dependencies: T018
- [ ] T020 Execute manual verification steps from `specs/005-implement-the-description/quickstart.md` (dry-run commands, inspect outputs) and log outcomes in PR notes, capturing the timing summary output for reference.
  - Dependencies: T017, T018, T019

## Dependencies

- Tests (T003–T008) must be created and observed failing before implementation tasks (T009–T014).
- Integration tasks (T015–T016) build on the completed helper implementations (T009–T014) and CLI wiring (T014).
- Timing instrumentation (T008) underpins the runtime summary surfaced by the CLI.
- Polish tasks (T017–T020) run only after all core and integration work is complete.

## Parallel Execution Example

```bash
task T003
task T004
task T005
task T007
task T008
```
