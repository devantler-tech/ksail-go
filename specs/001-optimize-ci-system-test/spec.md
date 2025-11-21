# Feature Specification: Optimize CI System Test Build Time

**Feature Branch**: `001-optimize-ci-system-test`  
**Created**: 2025-11-16  
**Status**: Implemented
**Input**: User description: "[chore]: Optimize CI system-test build time (currently 3m 26s)"

> **✅ IMPLEMENTATION STATUS**: This feature has been **SUCCESSFULLY IMPLEMENTED** with cache-based distribution. Current CI workflow implementation:
> - `.github/workflows/ci.yaml` contains `build-artifact` job that builds binary once per workflow
> - `.github/actions/prepare-ksail-binary/` composite action handles binary caching and restoration
> - `system-test` matrix jobs reuse cached binary via the prepare-ksail-binary action
> - Go module dependencies are cached using `actions/setup-go@v6` with `cache: true`
> - Implementation evolved to use cache-only distribution instead of CI artifacts
>
> **Note:** The specification below uses "artifact" terminology throughout (e.g., FR-002, FR-003, FR-006, FR-009, SC-003). However, the final implementation evolved to use cache-only distribution without formal CI artifacts. For consistency, references to "artifact" should be interpreted as "cached binary" in the context of the actual implementation.
>
## Clarifications

### Session 2025-11-16

- Q: Should the performance optimization focus solely on the `system-test` job or on every job within `.github/workflows/ci.yaml`? → A: Rework every job in `.github/workflows/ci.yaml`.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Maintainer Receives Faster CI Feedback (Priority: P1)

A maintainer opens a pull request and relies on the full CI workflow to validate build, lint, unit, and system-test stages. The workflow builds the binary once, shares it across every job that requires it, and finishes the longest matrix combinations in under two minutes, delivering consolidated feedback significantly faster than before.

**Why this priority**: Fast feedback protects contributor productivity and prevents long queues on the shared runners. Without this improvement, pull requests remain blocked across multiple CI stages for extended periods.

**Independent Test**: Trigger the workflow on a pull request with no code changes and verify that only one build step executes, shared across all dependent jobs, and that the system-test matrix completes within the target runtime.

**Acceptance Scenarios**:

1. **Given** a pull request with the full CI workflow enabled, **When** the workflow runs, **Then** a dedicated build job executes exactly once and the resulting binary artifact is available to every downstream job that needs it without further compilation.
2. **Given** no source or dependency changes between workflow re-runs, **When** any Go-based job executes, **Then** Go module dependencies resolve from cache with total download time under 15 seconds per job.
3. **Given** the workflow completes, **When** the maintainer inspects job durations, **Then** each system-test matrix job reports a run time of 1 minute 45 seconds or less without causing other jobs to exceed their pre-change durations by more than 10%.

---

### User Story 2 - Maintainer Diagnoses a Failing CI Job (Priority: P2)

A maintainer investigates a failing CI job—whether build, lint, unit, or system-test. They can inspect the shared binary metadata, cached dependency keys, and job logs to confirm the failure stems from test logic rather than an outdated binary or stale cache.

**Why this priority**: Troubleshooting confidence preserves trust in the optimized workflow and ensures failures remain actionable.

**Independent Test**: Force a test failure in one matrix entry and verify that diagnostics include artifact lineage and cache key details without requiring a local rebuild.

**Acceptance Scenarios**:

1. **Given** any CI job fails, **When** the maintainer opens the workflow summary, **Then** the job exposes the artifact version (commit SHA and build timestamp) and cache key used.
2. **Given** a dependency change updates `go.sum`, **When** the workflow runs, **Then** every job affected by Go modules invalidates the previous cache and downloads the new modules before executing tests or builds.

---

### User Story 3 - Contributor Introduces a New CI Job or Scenario (Priority: P3)

A contributor adds a new matrix entry or adjusts CI job arguments. The workflow automatically consumes the shared binary and caches without extra configuration, ensuring the new scenario benefits from the same performance gains.

**Why this priority**: Scaling the matrix should not reintroduce redundant builds or manual setup.

**Independent Test**: Add a temporary matrix entry in a draft pull request and verify that the job downloads the shared binary and completes within the target runtime without extra build logic.

**Acceptance Scenarios**:

1. **Given** a new matrix entry or CI job is added, **When** the workflow runs, **Then** the new job restores the shared binary from cache (if available) instead of compiling the binary again, falling back to a local build only if the cache is missed.
2. **Given** the workflow definition changes, **When** the workflow executes, **Then** caching behavior automatically applies with no additional keys or manual overrides required.

---

### Edge Cases

- How does the workflow behave when the shared binary build step fails or the artifact upload is unavailable? All downstream jobs must fail fast without attempting redundant builds.
- What happens when the cache is cold because of dependency or Go version changes? Jobs must fall back to downloading fresh modules and still complete successfully, even if slower.
- How are concurrent workflow runs handled when multiple pull requests execute simultaneously? Artifact names and cache keys must remain unique per run to avoid cross-run contamination across every job.
- What occurs if an individual matrix job requires platform-specific binaries in the future? Document how to fork the workflow while preserving reuse for compatible jobs.
- How do jobs that do not require the shared binary behave? They must skip cache restoration steps for the shared binary and only restore caches relevant to their own execution, while still benefiting from other caching improvements.

## Requirements *(mandatory)*

### Assumptions

- All current system-test matrix combinations run on the same operating system and can share a single compiled binary without compatibility issues.
- The CI platform retains artifacts long enough for downstream jobs within the same workflow run to download them reliably.
- Available cache storage is sufficient to store Go module dependencies for at least the most recent successful runs.
- Workflow maintainers have permission to modify CI configuration files and monitor run metrics.
- Baseline timing data (current 3 minutes 26 seconds per job) is available for comparison after implementation.
- Existing CI jobs that do not require the KSail binary can be adapted to skip artifact steps without altering their functional purpose.

### Functional Requirements

- **FR-001**: The CI workflow MUST introduce a dedicated build stage that runs exactly once per workflow execution before any downstream job requiring the KSail binary begins.
- **FR-002**: The dedicated build stage MUST produce the KSail binary with execution permissions and expose it as a versioned artifact scoped to the workflow run.
- **FR-003**: Every CI job that requires the KSail binary MUST download and execute the shared artifact instead of invoking a rebuild step.
- **FR-004**: The workflow MUST implement dependency caching for Go modules keyed by the Go toolchain version and dependency lock files so that all Go-based jobs benefit from cache hits on reruns without dependency changes.
- **FR-005**: Cache invalidation MUST occur automatically whenever Go version inputs or dependency manifests change, preventing stale dependencies from being reused by any job.
- **FR-006**: The workflow MUST fail any downstream job immediately if the shared artifact download or cache retrieval is unavailable, avoiding partial execution with inconsistent binaries.
- **FR-007**: The workflow configuration MUST retain all existing jobs, matrix combinations, and test commands, ensuring no reduction in coverage or test fidelity.
- **FR-008**: The workflow MUST capture per-job duration metrics (for example, via workflow summary output) so maintainers can compare post-change performance with the baseline for every job.
- **FR-009**: Developer-facing documentation (inline workflow comments or contributor docs) MUST describe the artifact-sharing model and cache invalidation rules, including guidance for jobs that do not require the shared binary.

### Key Entities

- **CI Workflow Run**: Represents a full automation execution for a commit or pull request, including the build stage, system-test matrix jobs, cache keys, artifact identifiers, and duration metrics.
- **System Test Artifact**: The compiled KSail binary packaged during the build stage, identified by the source commit and consumed by every matrix job within the same workflow run.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Average duration per system-test matrix job is 1 minute 45 seconds or less across three consecutive post-implementation runs on mainline commits.
- **SC-002**: Total wall-clock time for the entire CI workflow (all jobs in `.github/workflows/ci.yaml`) is reduced by at least 40% compared to the pre-change baseline, not exceeding 25 minutes on representative commits.
- **SC-003**: 100% of CI jobs that require the KSail binary consume the shared artifact produced by the single build stage, verified by workflow logs showing zero additional compile steps.
- **SC-004**: Cache hit rate for Go module downloads is at least 80% on reruns that do not modify dependency manifests.
- **SC-005**: No CI job that does not require the KSail binary experiences a runtime increase greater than 10% versus the pre-change baseline.
- **SC-006**: System-test pass rate remains unchanged from the pre-optimization baseline over the first ten runs after rollout, demonstrating no increase in flakiness.
