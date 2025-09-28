# Feature Specification: Workload Command Restructure for Reconcile, Apply, Install

**Feature Branch**: `004-move-description-ksail`
**Created**: September 28, 2025
**Status**: Draft
**Input**: User description: "Move `ksail reconcile` to `ksail workloads reconcile`, and add `ksail workloads apply`, and `ksail workloads install`.

Update: The namespace should be singular, exposing `ksail workload reconcile`, `ksail workload apply`, and `ksail workload install`.

None of the commands needs an implementation in this spec, I just want them to exist in the CLI so I can iterate on them in a later spec.

The reconcile command will be responsible for triggering reconciliation tools to sync workloads with a cluster. The apply command will wrap the `kubectl apply` command, to allow applying local files. The install command will wrap the `helm install` command to allow installing helm charts.

The reason for this is to structure align the CLI more with the expectations users might have. Supporting more ways to work with ones workloads allows flexibility, and supporting a larger set of use cases."

## Clarifications

### Session 2025-09-28

- Q: How should the legacy top-level `ksail reconcile` command behave once the workload namespace exists? → A: Remove; show command not found.
- Q: What should happen when someone runs `ksail workload apply` or `ksail workload install` before their full implementations land? → A: Show a "coming soon" message and exit with success (code 0).
- Q: How should `ksail workload reconcile` respond when no cluster context is configured? → A: Show a "coming soon" message and exit with success (code 0).
- Q: Should the namespace be named `workload` or `workloads`? → A: Use `workload` (singular) for all command invocations.

## User Scenarios & Testing *(mandatory)*

### Primary User Story

A platform engineer using the ksail CLI wants workload-related operations grouped together so they can reconcile, apply, or install workloads from a consistent namespace without remembering disparate top-level commands.

### Acceptance Scenarios

1. **Given** an initialized ksail project, **when** the user runs `ksail workload reconcile`, **then** the CLI must expose the command, show contextual help, and describe that it will sync workloads using the configured reconciliation tools.
2. **Given** a user with local Kubernetes manifests, **when** they invoke `ksail workload apply --help`, **then** the CLI must explain that the command applies local files (future behavior) and outline required inputs.
3. **Given** a user planning to deploy a Helm chart, **when** they run `ksail workload install --help`, **then** the CLI must present usage details indicating the intent to wrap Helm installations.

### Edge Cases

- What happens when a user runs `ksail workload reconcile` without having a cluster context configured?
- How does the CLI guide users if required parameters (e.g., manifest paths, chart references) are missing?
- The CLI must return a clear "command not found" style message directing users to `ksail workload reconcile` if they attempt the removed legacy top-level command.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The CLI MUST introduce a `workload` command group that is discoverable via `ksail --help` and provides a summary of workload-related actions.
- **FR-002**: The CLI MUST expose `ksail workload reconcile`, capturing the existing reconcile intent under the workload namespace and clarifying that it orchestrates syncing workloads with a cluster.
- **FR-002a**: Until cluster-awareness is implemented, `ksail workload reconcile` MUST fall back to a "coming soon" message and exit with success (code 0) when no cluster context is configured.
- **FR-003**: The CLI MUST expose `ksail workload apply` to position a future capability for applying local manifests (wrapping `kubectl apply`) and, until that iteration ships, MUST emit a "coming soon" message while exiting with success (code 0).
- **FR-004**: The CLI MUST expose `ksail workload install` to position a future capability for installing Helm charts (wrapping `helm install`) and, until that iteration ships, MUST emit a "coming soon" message while exiting with success (code 0).
- **FR-005**: The CLI MUST update command help and documentation so users see the workload namespace and its subcommands when requesting assistance or listing commands.
- **FR-006**: The CLI MUST remove the legacy top-level `ksail reconcile` command and surface a command-not-found style message that explicitly directs users to `ksail workload reconcile`.
- **FR-007**: The CLI MUST communicate that detailed behaviors for `apply` and `install` will arrive in a future iteration while avoiding broken or misleading execution paths.

## Review & Acceptance Checklist

> GATE: Automated checks run during main() execution

### Content Quality

- [x] No implementation details (languages, frameworks, APIs)
- [x] Focused on user value and business needs
- [x] Written for non-technical stakeholders
- [x] All mandatory sections completed

### Requirement Completeness

- [ ] No [NEEDS CLARIFICATION] markers remain
- [x] Requirements are testable and unambiguous
- [ ] Success criteria are measurable
- [x] Scope is clearly bounded
- [x] Dependencies and assumptions identified

---

## Execution Status

> Updated by main() during processing

- [x] User description parsed
- [x] Key concepts extracted
- [x] Ambiguities marked
- [x] User scenarios defined
- [x] Requirements generated
- [ ] Entities identified
- [ ] Review checklist passed

---
