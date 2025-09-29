# Feature Specification: KSail Cluster Provisioning Command

**Feature Branch**: `005-implement-the-description`
**Created**: 2025-09-28
**Status**: Draft
**Input**: User description: "Implement the `ksail cluster up` command to create a cluster based on the given ksail config. Make sure to support kind, k3d and eks distributions. The implementation must use existing packages."

## Clarifications

### Session 2025-09-28

- Q: When `ksail cluster up` is invoked and the target cluster already exists for the chosen distribution, what outcome should we guarantee? → A: Recreate only when `--force` flag is supplied, otherwise fail with an informative error
- Q: How are partial failures reported when infrastructure prerequisites (e.g., Docker daemon, cloud credentials) are unavailable? → A: Fail fast with clear, actionable error messages.
- Q: Do we wait for the provisioned cluster to become Ready before reporting success? → A: Yes, wait for Ready and kubeconfig usability.
- Q: If a required dependency (Docker daemon, AWS credentials) is missing at runtime, what should the command do? → A: Fail fast with a clear, actionable error.
- Q: After creating or verifying a cluster, how should the command manage kubeconfig? → A: Merge/update kubeconfig and set the new context as current.
- Q: What default timeout should apply while waiting for cluster readiness? → A: 5 minutes by default.

## User Scenarios & Testing *(mandatory)*

### Primary User Story

A platform engineer uses the KSail CLI to bootstrap a local or cloud Kubernetes environment. They run `ksail cluster up`, and the tool provisions a cluster that matches the distribution and configuration declared in their KSail project settings so they can begin deploying workloads immediately.

### Acceptance Scenarios

1. **Given** a project with a valid KSail configuration targeting a local distribution (Kind or K3d), **When** the engineer runs `ksail cluster up`, **Then** KSail provisions the cluster by creating the required containers inside the detected container engine (Docker or Podman) using the combined configuration from Viper. The user receives feedback on completion or actionable errors if provisioning fails.
2. **Given** a KSail configuration targeting the cloud EKS distribution and containing all required credentials or connection details, **When** the engineer runs `ksail cluster up`, **Then** the EKS cluster is created or brought to an active state using the combined configuration from Viper. The user receives feedback on completion or actionable errors if provisioning fails.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST load the active KSail project configuration before executing `ksail cluster up`, validating that a supported distribution is specified.
- **FR-002**: The command MUST provision or start a Kind cluster when the configuration declares the Kind distribution, applying user-defined settings such as cluster name and node parameters.
- **FR-003**: The command MUST provision or start a K3d cluster when the configuration declares the K3d distribution, honoring the configuration values relevant to K3d environments.
- **FR-004**: The command MUST provision or start an EKS cluster when the configuration declares the EKS distribution, using the provided account, region, and cluster sizing information.
- **FR-005**: The command MUST reuse existing KSail provisioning capabilities rather than introducing new distribution frameworks, ensuring consistent behavior with current lifecycle operations.
- **FR-006**: The command MUST surface success using the existing provisioner/notify output streams, appending a concise timing summary that highlights both the overall command duration and the duration of individual steps. On failure, it MUST provide an actionable remediation hint (for example, "Start Docker and rerun the command") and confirm that no orphaned infrastructure remains when cleanup is feasible.
- **FR-007**: The command MUST detect when the target cluster already exists, treating the operation as a readiness verification unless the user passes `--force`, in which case the cluster is recreated.
- **FR-008**: The command MUST block until the provisioned cluster reports Ready status and the kubeconfig is usable before signaling success for Kind, K3d, and EKS distributions.
- **FR-009**: The command MUST perform dependency checks and fail fast with explicit remediation guidance when prerequisites such as Docker daemons or cloud credentials are unavailable.
- **FR-010**: The command MUST merge or update the user's kubeconfig with the cluster connection details and set the new cluster context as current upon successful completion.
- **FR-011**: The command MUST enforce a default 5-minute timeout when waiting for cluster readiness, after which it fails with guidance unless the user overrides the timeout.

### Key Entities

- **KSail Project Configuration**: Defines the desired distribution (Kind, K3d, or EKS), cluster naming, sizing, and provider-specific options that drive provisioning behavior.
- **Provisioning Response**: Represents the outcome of the cluster creation attempt, including success confirmation, contextual metadata (cluster name, endpoint), and any error messages presented to the user.

## Review & Acceptance Checklist

### Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness

- [x] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [x] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

## Execution Status

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [x] Entities identified
- [x] Review checklist passed
