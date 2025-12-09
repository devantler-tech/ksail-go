# Tasks: Flux OCI Integration for Automated Reconciliation

**Input**: Design documents from `specs/001-flux-oci-integration/`
**Prerequisites**: plan.md, spec.md (user stories), research.md, data-model.md, contracts/

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Ensure repo-level tooling and environment are ready for Flux + OCI work.

- [x] T001 Verify Go toolchain and linters for KSail-Go in repository root
- [x] T002 [P] Confirm KSail-Go builds successfully with `go build ./...` from `src/`
- [x] T003 [P] Confirm KSail-Go tests pass with `go test ./...` from `src/`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core plumbing and config extensions required before any user story implementation.

- [x] T004 Extend KSail-Go cluster configuration model in `pkg/apis` to support `GitOpsEngine`, registry, and Flux interval fields
- [x] T005 Update config parsing and validation in `pkg/cmd` / `pkg/di` to read and validate Flux + registry settings
- [x] T006 [P] Add data structures for `OCIRegistry`, `OCIArtifact`, and Flux CR models in `pkg/apis`
- [x] T007 [P] Ensure existing DI wiring can inject new services via interfaces in `pkg/svc`

**Checkpoint**: Foundation ready – user stories can now be implemented.

---

## Phase 3: User Story 1 - Bootstrap Cluster with Flux (Priority: P1) 🎯 MVP

**Goal**: Install Flux controllers during cluster bootstrap or on-demand when `GitOpsEngine: Flux` is configured.

**Independent Test**: Create a cluster with Flux enabled and verify Flux controllers and CRDs are present and healthy.

### Implementation for User Story 1

- [x] T008 [P] [US1] Define `FluxInstaller` interface and types in `pkg/svc/installer/flux` (install/uninstall/status)
- [x] T009 [P] [US1] Implement `FluxInstaller` using Flux Installer Go SDK in `pkg/svc/installer/flux`
- [x] T010 [US1] Integrate `FluxInstaller` into cluster bootstrap path in `cmd/cluster` (enable via `GitOpsEngine: Flux`)
- [x] T011 [US1] Add logic to perform on-demand Flux install for existing clusters in `cmd/cluster`
- [x] T012 [US1] Add basic unit tests for `FluxInstaller` and cluster bootstrap integration in `pkg/svc/installer/flux` and `cmd/cluster` (write tests before implementation where practical)
- [x] T013 [US1] Implement detection logic to handle cases where Flux is already installed and avoid duplicate installation attempts

**Checkpoint**: Creating a cluster with Flux enabled installs controllers and CRDs without affecting other workloads.

---

## Phase 4: User Story 2 - Provision Local OCI Registry (Priority: P1)

**Goal**: Provision a localhost-only, unauthenticated OCI registry (`registry:3`) with persistent storage accessible from both cluster and workstation.

**Independent Test**: Provision the registry and verify image push/pull succeeds from host and from pods in the cluster.

### Implementation for User Story 2

- [x] T014 [P] [US2] Define `RegistryService` interface and types in `pkg/svc/provisioner/registry` (create/start/stop/status)
- [x] T015 [P] [US2] Implement `RegistryService` for `registry:3` using container engine interaction in `pkg/svc/provisioner/registry`
- [x] T016 [US2] Wire `RegistryService` into cluster creation/update commands in `cmd/cluster` when registry is enabled
- [x] T017 [US2] Ensure registry endpoint is bound to `localhost:<port>` and persisted via volume in `pkg/svc/provisioner/registry`
- [x] T018 [US2] Add unit tests for `RegistryService` behavior and basic CLI integration tests for enabling the registry, including verifying push from host and pull from pods
- [x] T019 [US2] Implement cleanup of registry volumes and associated resources on cluster delete in `pkg/svc/provisioner/registry` to satisfy FR-015

**Checkpoint**: Local registry can be provisioned and used for manual image push/pull.

---

## Phase 5: User Story 3 - Build and Push OCI Artifacts (Priority: P2)

**Goal**: Build OCI artifacts from Kubernetes manifest directories and push them to the local registry with semantic versioning.

**Independent Test**: Build artifacts from sample manifests, push to registry, and verify artifact metadata and contents.

### Implementation for User Story 3

- [x] T020 [P] [US3] Define `WorkloadArtifactBuilder` interface and types in `pkg/workload/oci`
- [x] T021 [P] [US3] Implement `WorkloadArtifactBuilder` using `google/go-containerregistry` in `pkg/workload/oci`
- [ ] T022 [US3] Add CLI command (e.g., `ksail workload build`) in `cmd/workload` that uses `WorkloadArtifactBuilder`
- [ ] T023 [US3] Enforce semantic versioning for artifact tags and validate source manifest directories
- [x] T024 [US3] Implement structural validation of OCI artifacts before push (e.g., required labels, manifest presence) in `pkg/workload/oci` to satisfy FR-017
- [ ] T025 [US3] Add unit tests and snapshot tests for artifact building, validation, and CLI output in `pkg/workload/oci` and `cmd/workload` (tests first where practical)

**Checkpoint**: OCI artifacts can be built and pushed to the local registry and listed with correct versions.

---

## Phase 6: User Story 4 - Configure Flux Custom Resources (Priority: P2)

**Goal**: Generate and apply Flux `OCIRepository` and `Kustomization` resources that track OCI artifact repositories and apply manifests.

**Independent Test**: Create Flux resources for a sample artifact and verify Flux fetches metadata and applies manifests.

### Implementation for User Story 4

- [x] T026 [P] [US4] Implement helpers in `pkg/svc/installer/flux` to generate `OCIRepository` and `Kustomization` manifests from config
- [ ] T027 [P] [US4] Add a CLI command or subcommand in `cmd/workload` or `cmd/cluster` to generate/apply these Flux resources
- [x] T028 [US4] Ensure default reconciliation interval of 1 minute is applied when not overridden
- [x] T029 [US4] Validate that generated resources use the `oci://localhost:<port>/<project-name>` pattern and no auth
- [x] T030 [US4] Add unit tests to validate manifest generation and basic integration tests applying them to a dev cluster

**Checkpoint**: Flux can track and apply manifests from local OCI repositories as configured by KSail-Go.

---

## Phase 7: User Story 5 - Trigger Reconciliation on Artifact Updates (Priority: P1)

**Goal**: Ensure that when new artifact versions are pushed, Flux detects the changes and reconciles cluster state automatically.

**Independent Test**: Push a new artifact version and observe Flux applying changes within the reconciliation interval.

### Implementation for User Story 5

- [ ] T031 [P] [US5] Confirm Flux reconciliation interval configuration in `OCIRepository`/`Kustomization` resources supports frequent polling
- [ ] T032 [US5] Implement helper in `pkg/svc/installer/flux` to check status/conditions for reconciliation results
- [ ] T033 [US5] Add a simple CLI status command in `cmd/workload` or `cmd/cluster` to surface high-level success/failure based on Flux CRs
- [ ] T034 [US5] Add tests simulating an artifact update and verifying Flux picks up the new version via status inspection

**Checkpoint**: End-to-end automated reconciliation from new artifact push to applied resources is demonstrable.

---

## Phase 8: User Story 6 - Manual Reconciliation via CLI (Priority: P3)

**Goal**: Provide a `ksail workload reconcile` command that triggers immediate Flux reconciliation without waiting for the polling interval.

**Independent Test**: Run the reconcile command and verify Flux immediately reconciles configured sources.

### Implementation for User Story 6

- [ ] T035 [P] [US6] Implement reconciliation trigger helper in `pkg/svc/installer/flux` (e.g., patching annotations or using Flux APIs)
- [ ] T036 [US6] Implement `ksail workload reconcile` command in `cmd/workload` that calls the helper
- [ ] T037 [US6] Ensure command returns appropriate exit codes based on reconciliation success/failure
- [ ] T038 [US6] Add unit tests and snapshot tests for `ksail workload reconcile` behavior and output (tests first where practical)

**Checkpoint**: Operators can force immediate reconciliation for debugging and rapid iteration.

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Consolidate docs, improve UX, and handle storage/pruning concerns.

- [ ] T036 [P] Update KSail-Go docs under `docs/` to describe Flux + OCI flows and quickstart steps
- [ ] T039 Review error messages across Flux and registry services for clarity and consistency, including common edge cases (registry unavailable, corrupted artifact, auth failure)
- [ ] T040 Implement basic artifact pruning options for local registry in `pkg/svc/provisioner/registry` (for FR-018)
- [x] T041 [P] Add or refine end-to-end/system tests in existing CI workflows for the Flux + OCI path, including at least one edge-case scenario
- [ ] T042 Run through `quickstart.md` step-by-step and fix any mismatches in commands or behavior
- [x] T043 Run `golangci-lint run` and `go test ./...` to confirm feature-level quality gates before merge

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies – can start immediately.
- **Foundational (Phase 2)**: Depends on Phase 1 completion – blocks all user stories.
- **User Stories (Phases 3–8)**: All depend on Phase 2; then can proceed in story priority order (P1 → P2 → P3), or in parallel if capacity allows.
- **Polish (Phase 9)**: Depends on all targeted user stories being complete.

### Parallel Opportunities

- Tasks marked [P] can be implemented in parallel as they touch separate files or are loosely coupled.
- After Phase 2, implementation of different user stories (US1–US6) can proceed in parallel by different contributors, as each story is independently testable.
