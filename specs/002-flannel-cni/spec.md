# Feature Specification: Flannel CNI Support

**Feature Branch**: `002-flannel-cni`
**Created**: 2025-11-15
**Status**: Draft
**Input**: User description: "Add Flannel CNI to ksail to allow users to specify and use Flannel as their chosen CNI."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Select Flannel During Init (Priority: P1)

A platform engineer initializing a new local Kubernetes project chooses "Flannel" as the CNI option during the interactive or flag-based `init` flow and then creates the cluster. The cluster comes up successfully with Flannel pods running and all nodes reach Ready without requiring any manual manifest edits.

**Why this priority**: Core value of the feature is enabling an alternative lightweight CNI. Without this story the feature provides no user-facing benefit.

**Independent Test**: Run project initialization specifying Flannel; create cluster; verify nodes Ready and Flannel DaemonSet pods all Available. No other functionality needed to demonstrate value.

**Acceptance Scenarios**:

1. **Given** a fresh project with no cluster, **When** the user runs initialization selecting Flannel, **Then** project configuration records Flannel as chosen CNI.
2. **Given** a project configured for Flannel, **When** the user creates the cluster, **Then** all Flannel control-plane and daemon pods reach Ready and nodes become Ready within the success criteria timing.

---

### User Story 2 - View CNI Status (Priority: P2)

An engineer queries cluster status and the CLI clearly states that Flannel is the active CNI along with a summary (version, readiness) so they can confirm the environment matches expectations.

**Why this priority**: Visibility reduces misconfiguration and accelerates troubleshooting; secondary to initial selection.

**Independent Test**: After cluster creation, execute status command; output includes Flannel identification & health metrics independent of other features.

**Acceptance Scenarios**:

1. **Given** a running cluster created with Flannel, **When** the user runs a status/info command, **Then** output displays CNI="Flannel" plus readiness summary (e.g., all pods ready count) and no conflicting CNI.

---

### User Story 3 - Prevent Conflicting CNI (Priority: P3)

If a user attempts to select Flannel while another CNI (e.g., Cilium) is already installed or specified, the system warns and blocks conflicting installation, guiding the user toward a clean re-initialization or migration process.

**Why this priority**: Avoids broken networking from overlapping CNIs; tertiary because it protects reliability after core selection and visibility exist.

**Independent Test**: Attempt to install Flannel on a cluster already using a different CNI; verify system blocks and provides remediation options without needing other feature components.

**Acceptance Scenarios**:

1. **Given** a cluster previously initialized with another CNI, **When** user tries to switch directly to Flannel without teardown, **Then** the system aborts with a clear conflict message.

---

Additional stories (lower priority) may include: Optional backend mode selection; metrics reporting enhancements.

### Edge Cases

- User selects Flannel on a distribution outside supported set (Kind, K3d) → graceful error with guidance to use a supported distribution or alternative CNI.
- Attempt to enable network policies expecting enforcement (Flannel does not provide native policy) → warning explaining limitation and suggesting a policy-capable CNI (e.g., Cilium) if needed.
- User re-runs init with different CNI after cluster exists → instructed that full cluster recreation is required; in-place migration not supported.
- Missing node readiness within timeout → surface diagnostic suggesting checking Flannel pod logs and verifying node network interfaces.
- Selecting unsupported backend mode (only vxlan is available) → validation error clarifying vxlan is the fixed mode for this release.
- Cluster already has residual CNI resources from failed installation → detection and abort with cleanup instructions (tear down cluster or manually remove artifacts before retry).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to select Flannel as a CNI option during project initialization via flag or interactive prompt.
- **FR-002**: System MUST record the selected CNI (Flannel) in project configuration for subsequent lifecycle commands.
- **FR-003**: System MUST install Flannel prior to workloads that depend on stable pod networking.
- **FR-004**: System MUST validate post-installation readiness (all Flannel pods Ready) within a default timeout of 120s; failure MUST include guidance (e.g., check DaemonSet logs).
- **FR-005**: System MUST present active CNI="Flannel" in status/inspection output including readiness counts and backend mode.
- **FR-006**: System MUST prevent installation if another CNI is already active unless a documented migration path is followed (not in scope here).
- **FR-007**: System MUST warn users that Flannel does not provide native network policy enforcement and clarify security implications.
- **FR-008**: System MUST fail fast with clear error if user selects Flannel on a distribution not supported (supported: Kind, K3d).
- **FR-009**: System MUST fix backend mode to `vxlan` (no user selection in this scope) and document this constraint.
- **FR-010**: System MUST ensure no residual conflicting CNI resources remain prior to Flannel activation (basic conflict detection & abort).
- **FR-011**: System MUST instruct users that switching from another CNI to Flannel requires full cluster recreation; in-place migration is not supported.
- **FR-012**: System MUST expose Flannel version (image tag) in status output for traceability.
- **FR-013**: System SHOULD provide a diagnostic command or flag to dump Flannel pod events when readiness fails.

### Key Entities

- **Cluster Configuration**: Holds chosen CNI type (now includes value `Flannel`) and optional backend mode (if enabled). Attributes: `cniType`, `cniOptions` (may include `backendMode`).
- **CNI Status Summary**: Represents current CNI label, readiness counts, and warnings (e.g., network policy limitation). Attributes: `name`, `podsReady`, `warnings`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: User can initialize and create a local cluster with Flannel selected (Kind or K3d) in under 5 minutes total (init start → all nodes Ready).
- **SC-002**: 95% of clusters created with Flannel reach node Ready state within 120 seconds after CNI deployment (under normal local resource conditions); 100% within 180 seconds.
- **SC-003**: Status command correctly reports Flannel as active CNI with accurate readiness and version in 100% of test runs across Kind and K3d.
- **SC-004**: Attempting conflicting CNI installation produces a clear explanatory error in < 2 seconds with remediation guidance.
- **SC-005**: At least 90% of regression test runs (Kind + K3d matrix) succeed without CNI-related failures after introduction of Flannel support.
- **SC-006**: Users receive exactly one warning about lack of network policy enforcement when choosing Flannel (verified in interactive and flag flows).
- **SC-007**: Diagnostic flag outputs pod events within 2 seconds on failure scenarios.

## Non-Functional Requirements

- **NFR-Performance**: Flannel installation step SHOULD complete (pods Ready) within 60s on a 2 CPU / 4GB RAM dev machine.
- **NFR-Reliability**: Failure modes MUST provide actionable messages (missing readiness, conflicting CNI) with exit codes ≠ 0.
- **NFR-Observability**: Status output MUST include version and readiness counts; diagnostics MUST include last 20 Flannel events.
- **NFR-Security**: No network policy support—warning MUST appear before cluster creation completes; no privileged escalation beyond Flannel defaults.
- **NFR-Compatibility**: Supports Kind v0.23+ and K3d v5+; error if versions older.
- **NFR-Documentation**: Help text MUST mention vxlan fixation and lack of policy enforcement.

## Assumptions

- Backend mode fixed to `vxlan` for initial implementation.
- Network policy enforcement is out of scope for Flannel; users needing policies should select a different CNI.
- Supported distributions limited to Kind and K3d for this release (EKS excluded due to managed networking requirements).
- Switching CNIs requires full cluster recreation; no live migration path.

## Resolved Clarifications

1. Supported distributions: Kind, K3d only.
2. Backend mode: Fixed `vxlan` (no user selection).
3. Migration approach: Full cluster recreation required to switch from another CNI.
