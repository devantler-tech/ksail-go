# Tasks: Workload Command Restructure for Reconcile, Apply, Install

**Input**: Design documents from `/Users/ndam/git-personal/monorepo/projects/ksail-go/specs/004-move-description-ksail/`
**Prerequisites**: `plan.md` (required), `research.md`, `data-model.md`, `contracts/workload-cli.md`, `quickstart.md`

## Phase 3.1: Setup

- [ ] T001 Run baseline test suite with `go test ./...` to confirm the workspace is clean before introducing new workload commands.

## Phase 3.2: Tests First (TDD) ⚠️ MUST COMPLETE BEFORE 3.3

**Critical:** These test additions should be committed while still failing (or not compiling) until implementation lands.

- [ ] T002 [P] Extend `cmd/workload/workload_test.go` with snapshot-based help coverage for the workload namespace, asserting `ksail workload --help` plus `ksail workload reconcile|apply|install --help` outputs against new fixtures under `cmd/__snapshots__/`.
- [ ] T003 [P] In the same `cmd/workload/workload_test.go`, add placeholder command behavior tests to ensure `reconcile`, `apply`, and `install` emit the "Coming soon" message via `notify.Infoln` and exit with code 0.
- [ ] T004 [P] Extend `cmd/root_test.go` with a regression test that running `ksail reconcile` surfaces Cobra's unknown-command error plus the guidance string from the workload migration contract.

## Phase 3.3: Core Implementation (ONLY after tests are failing)

- [ ] T005 Implement the workload command group in `cmd/workload/workload.go`, providing `NewWorkloadCmd()` with rich help text and wiring it to internal constructors for each subcommand.
- [ ] T006 Flesh out the workload subcommands in `cmd/workload/workload.go` (or supporting files) so `reconcile`, `apply`, and `install` each print their contract-defined "Coming soon" message using `notify.Infoln` and return success.
- [ ] T007 Update `cmd/root.go` to register `NewWorkloadCmd()`, remove the legacy top-level reconcile registration, and adjust help ordering to surface the new namespace.
- [ ] T008 Remove or repurpose the obsolete `cmd/reconcile.go` so no standalone root-level command remains (delete the file or convert it to a thin shim invoking the workload command constructor as appropriate).
- [ ] T009 Enhance `runWithArgs` in `main.go` (or the closest invocation boundary) to intercept the unknown-command error for `reconcile` and append the guidance string directing users to `ksail workload reconcile`.

## Phase 3.4: Integration & Validation

- [ ] T010 [P] Re-record help snapshots after implementation by running `go test ./cmd -run TestWorkloadHelp -update` and committing the generated `cmd/__snapshots__/` entries.
- [ ] T011 Format updated Go sources with `gofmt`/`goimports`, covering `main.go`, `cmd/root.go`, and the new `cmd/workload` package.
- [ ] T012 Run `go test ./...` to ensure the full suite (including new workload tests) passes.
- [ ] T013 [P] Execute `golangci-lint run --timeout 5m` from the repository root to satisfy constitutional lint requirements.
- [ ] T014 [P] Walk through the quickstart validation steps in `quickstart.md`, confirming binary build, command help listings, placeholder outputs, and the legacy guidance message.
- [ ] T015 Update `quickstart.md` (and any affected docs) to surface the new workload namespace, subcommands, and "Coming soon" placeholders so user-facing guidance reflects FR-005 expectations.
- [ ] T016 Review coverage output from `go test ./...` to ensure new packages maintain >90% project coverage; investigate and shore up tests if the threshold dips.

## Dependencies

- T002–T004 must complete (and fail) before starting T005–T009.
- T005 precedes T006 because the subcommand constructors hang off `NewWorkloadCmd()`.
- T007 depends on successful wiring from T005–T006.
- T008 follows T007 so the root command no longer references the removed file.
- T009 depends on the root wiring changes from T007 and must run before validation tasks.
- T010–T014 run only after all core implementation tasks complete.
- T015 depends on T005–T009 so documentation reflects final command surface; run it before closing validation.
- T016 executes after T012 to confirm constitutional coverage expectations.

## Parallel Example

Launch these test-authoring tasks together before implementation:

```text
/task run T002
/task run T003
/task run T004
```

## Notes

- Marked [P] tasks target separate files or command-line validation and can execute concurrently when their dependencies are satisfied.
- Ensure failing tests and snapshots are committed prior to implementation, aligning with the constitution’s TDD mandate.
- Snapshot updates (T010) should be the only point where `-update` flags are used so earlier failing tests retain their guard rails.
- Follow the repository convention of one test file per source file—`cmd/workload/workload.go` must pair only with `cmd/workload/workload_test.go`.
