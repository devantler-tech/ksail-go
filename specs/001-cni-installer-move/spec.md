# Feature Specification: CNI Installer Consolidation

**Feature Branch**: `[001-cni-installer-move]`
**Created**: 2025-11-14
**Status**: Draft
**Input**: User description: "Implement the feature specification based on the updated constitution. I want to move all CNI installer packages and helpers to pkg/svc/installer/cni/"

## Clarifications

### Session 2025-11-14

- Q: When the package relocation breaks runtime behavior (e.g., missing Helm chart, CNI install failure), how should the system handle rollback or recovery? → A: Preserve existing error handling behavior unchanged; rely on Git revert if relocation breaks builds
- Q: What Helm version compatibility must be maintained for CNI installers after the relocation? → A: Helm 3.8.0+ (matches current CI/dev environment)
- Q: Should the relocation introduce any new logging or metrics to track CNI installer usage patterns, or strictly preserve existing observability? → A: Preserve existing notify/timer output exactly; no new logging or metrics
- Q: Should the old package paths (pkg/svc/installer/calico, pkg/svc/installer/cilium) be immediately deleted after relocation, or maintained as deprecated aliases for a transition period? → A: Delete immediately after updating all imports; no transition period needed
- Q: Should CNI installer security posture be validated after relocation (e.g., Helm chart signature verification, RBAC requirements)? → A: Preserve existing security mechanisms unchanged; validate no regressions introduced
- Q: Should CNI installers be consolidated directly under pkg/svc/installer/cni or maintain subdirectories? → A: Maintain subdirectories (pkg/svc/installer/cni/cilium/, pkg/svc/installer/cni/calico/); only shared helpers (CNIInstallerBase, utilities) move to pkg/svc/installer/cni/ root

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Maintain installer functionality after package move (Priority: P1)

Platform engineers running `ksail cluster up` expect Cilium/Calico CNIs to install successfully even after the code lives under `pkg/svc/installer/cni/`.

**Why this priority**: Preserving working cluster bootstrap flows is critical; any regression blocks developers from creating local environments.

**Independent Test**: Automated unit tests covering `pkg/svc/installer/cni` execute successfully after the move, and an integration smoke test verifies installers still succeed without relying on old import paths.

**UX Consistency Notes**: No CLI surface changes; existing notify/timer output from cluster commands must remain verbatim because only internal package paths shift.

**Performance Budget**: CNI installation must continue to complete within existing timeout defaults (≤ 10 minutes) and still emit progress via existing timers/logs.

**Acceptance Scenarios**:

1. **Given** a repository using the new package layout, **When** automated unit tests targeting the CNI installer packages run, **Then** they pass without import errors.
2. **Given** a developer bootstraps a Kind cluster with default CNI, **When** `ksail cluster up` runs, **Then** Cilium installs successfully with no CLI output changes beyond timestamps.

---

### User Story 2 - Simplify adding new CNIs (Priority: P2)

Contributors adding a new CNI should be able to place installers and shared helpers under `pkg/svc/installer/cni/` without touching unrelated directories.

**Why this priority**: Consolidating shared logic reduces duplication and clarifies extension points for future CNIs.

**Independent Test**: Scaffolding a sample installer that imports the new `cni` helpers should compile and reuse the shared base without referencing the root `installer` package directly.

**UX Consistency Notes**: Developer-facing docs (e.g., contribution guide) must reflect the new layout so authors follow consistent notify/timer usage.

**Performance Budget**: Build/test cycles (go build/test) must remain under existing norms (< 2 minutes for affected packages) since only file moves occur.

**Acceptance Scenarios**:

1. **Given** the restructured layout, **When** a developer creates a stub installer under `pkg/svc/installer/cni/example/`, **Then** imports resolve using the shared CNI helpers from `pkg/svc/installer/cni` without referencing legacy paths.

---

### User Story 3 - Maintain upgrade path for existing imports (Priority: P3)

Internal packages that referenced `pkg/svc/installer/calico` or `.../cilium` must build once they switch to the new paths, ensuring a clear migration path.

**Why this priority**: Guarantees dependent services/code can adopt the new structure without churn or runtime surprises.

**Independent Test**: The standard repository build pipeline completes successfully after updating imports, confirming no lingering references to deleted paths.

**UX Consistency Notes**: No user-facing messaging changes; release notes should mention the internal path rename for maintainers.

**Performance Budget**: Full repository build should remain within current CI bounds (< 5 minutes); relocation must not add heavy operations.

**Acceptance Scenarios**:

1. **Given** all references updated, **When** automated static analysis runs, **Then** no lint errors report missing packages or unused files due to the move.

---

### Edge Cases

- Imports forgotten in rarely used packages or tests cause build failures—ensure tooling detects these by running `go build ./...` and `go test ./...`.
- Local scripts or documentation referencing old paths need updates; otherwise contributors may follow stale guidance.
- **Failure Recovery**: System preserves existing error handling behavior unchanged; if relocation breaks builds or runtime behavior, recovery relies on Git revert to restore working state. No automated rollback mechanism introduced.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST relocate `CNIInstallerBase`, helper functions, and readiness utilities into `pkg/svc/installer/cni/base.go` (single file at root of cni package) while preserving exported APIs.
- **FR-002**: System MUST move Cilium and Calico installer packages to `pkg/svc/installer/cni/cilium/` and `pkg/svc/installer/cni/calico/` subdirectories respectively, updating their module names accordingly.
- **FR-003**: System MUST update all imports, mocks, and tests to reference the new paths (`pkg/svc/installer/cni/` for shared helpers, `pkg/svc/installer/cni/cilium/` and `pkg/svc/installer/cni/calico/` for installers), then immediately delete old package directories (`pkg/svc/installer/calico/`, `pkg/svc/installer/cilium/`, `pkg/svc/installer/cni_helpers.go`) with no transition period or deprecated aliases maintained.
- **FR-004**: System MUST ensure `go test ./pkg/svc/installer/cni/...` and `go test ./pkg/svc/installer/...` pass locally and in CI.
- **FR-005**: System MUST refresh contributor documentation or inline comments describing where new CNIs should live.

### Key Entities *(include if feature involves data)*

- **CNIInstallerBase**: Shared struct encapsulating Helm client, kubeconfig, and readiness waits used across CNIs.
- **HelmRepoConfig / HelmChartConfig**: Value objects controlling CNI Helm install parameters; remain unchanged but relocate under the CNI module.

### Quality & Performance Constraints *(align with Constitution)*

- **QC-001**: Unit tests MUST cover base installer helpers plus Calico/Cilium flows; add or relocate tests to maintain coverage after the move.
- **QC-002**: CLI output and observability MUST remain unchanged; preserve existing notify/timer output exactly with no new logging, metrics, or debug instrumentation introduced during relocation.
- **QC-003**: Build/test cycles MUST remain within existing limits (CNI package tests completing within 90 seconds); capture timing evidence when running validation steps.
- **QC-004**: Security mechanisms MUST remain unchanged; existing Helm installation behavior and RBAC requirements preserved exactly with validation that no security regressions are introduced during relocation.

## Assumptions & Dependencies

- Existing Helm installer interfaces remain unchanged; the refactor must integrate with current CNI installers without modifying Helm clients.
- **Helm Compatibility**: CNI installers require Helm 3.8.0 or later, matching the current CI/development environment.
- CI infrastructure can execute unit, lint, and smoke tests to validate the relocation.
- No new CNIs are introduced in this scope; effort strictly relocates existing implementations and shared helpers.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Automated unit tests for the relocated CNI installer packages complete with zero failures and within 90 seconds on a standard dev machine.
- **SC-002**: The repository build pipeline completes successfully with no import errors or lint violations referencing removed paths.
- **SC-003**: Contribution docs explicitly reference `pkg/svc/installer/cni/` as the canonical location for CNI installers.
- **SC-004**: CI build and lint stages pass on the feature branch without increased runtime beyond established baselines.
