# Tasks: KSail Cluster Provisioning Command

**Input**: Design documents from `/specs/005-implement-the-description/`
**Prerequisites**: plan.md, research.md, data-model.md, contracts/cluster-up.md, quickstart.md

## Phase 3.1: Setup

- [ ] T001 Create cluster up orchestration package skeleton (`pkg/clusterup/README.md`, `pkg/clusterup/doc.go`).
  - Dependencies: Implementation plan approved (`plan.md`)
  - Notes: Document package intent and export guidelines without adding functional code.
- [ ] T002 [P] Prepare shared test fixtures under `pkg/clusterup/testdata/` (kubeconfig templates, fake node payloads, AWS profile stubs).
  - Dependencies: T001
  - Notes: Provide reusable files referenced by dependency, readiness, and kubeconfig tests.

## Phase 3.2: Tests First (TDD)

- [ ] T003 [P] Author failing dependency-check tests in `pkg/clusterup/dependencies_test.go` covering Docker/Podman detection and AWS credential errors.
  - Dependencies: T001, T002
- [ ] T004 [P] Author failing readiness waiter tests in `pkg/clusterup/readiness_test.go` for success, timeout, and API failure flows using fake clientsets.
  - Dependencies: T001, T002
- [ ] T005 [P] Author failing kubeconfig management tests in `pkg/clusterup/kubeconfig_test.go` validating context creation, switching, and persistence to disk.
  - Dependencies: T001, T002
- [ ] T006 Author failing runner orchestration tests in `pkg/clusterup/runner_test.go` covering new cluster creation, reuse, force recreate, dependency failure, and readiness timeout.
  - Dependencies: T003, T004, T005
- [ ] T007 [P] Expand `cmd/cluster/up_test.go` contract coverage for CLI flag wiring, runner invocation, dependency error messaging, readiness timeout exit codes, and confirmation that success output routes through existing notify/provisioner helpers.
  - Dependencies: T006
- [ ] T008 [P] Add timing instrumentation tests in `pkg/clusterup/telemetry_test.go` (or within runner tests) to validate per-stage duration capture and summary formatting.
  - Dependencies: T003, T004, T005

## Phase 3.3: Core Implementation (after tests are failing)

- [ ] T009 Implement spec and distribution config loading helpers in `pkg/clusterup/config_loader.go` returning validated `v1alpha1.Cluster` and provider configs expected by runner tests.
  - Dependencies: T006
- [ ] T010 Implement dependency checking logic and `DependencyCheckResult` struct in `pkg/clusterup/dependencies.go` to satisfy T003 expectations.
  - Dependencies: T003
- [ ] T011 Implement readiness waiter with `ReadinessProbeConfig` in `pkg/clusterup/readiness.go` using client-go polling to satisfy T004.
  - Dependencies: T004
- [ ] T012 Implement kubeconfig management utilities in `pkg/clusterup/kubeconfig.go` ensuring context creation and persistence to satisfy T005.
  - Dependencies: T005
- [ ] T013 Implement `Runner` orchestration, provisioner factory wiring, and `ProvisioningOutcome` struct in `pkg/clusterup/runner.go` to satisfy T006.
  - Dependencies: T009, T010, T011, T012
- [ ] T014 Refactor `cmd/cluster/up.go` to invoke the runner, rely on existing notify/provisioner helpers for success messaging, and surface exit codes/errors expected by T007.
  - Dependencies: T013, T007

## Phase 3.4: Integration

- [ ] T015 Integrate runner injection seams for tests and command wiring (e.g., update `cmd/cluster/cluster.go` or supporting file) so CLI and benchmark harnesses can hook into the runner cleanly.
  - Dependencies: T013, T007
- [ ] T016 Connect runner factories to existing Kind, K3d, and EKS provisioners/config managers (e.g., `pkg/clusterup/factory.go`) ensuring real provisioning paths function.
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
- Runner integration (T015–T016) builds on the completed runner implementation (T013) and CLI wiring (T014).
- Timing instrumentation (T008) underpins the runtime summary surfaced by the runner and CLI.
- Polish tasks (T017–T020) run only after all core and integration work is complete.

## Parallel Execution Example

```bash
task T003
task T004
task T005
task T007
task T008
```
