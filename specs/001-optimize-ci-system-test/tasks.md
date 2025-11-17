# Tasks: Optimize CI System Test Build Time

**Input**: Design documents from `/specs/001-optimize-ci-system-test/`
**Prerequisites**: plan.md (required), spec.md (required for user stories), research.md, data-model.md, contracts/

**Tests**: Define failing verification steps (workflow dry-runs for the metrics script and composite action) before implementation; smoke tests remain embedded in consuming jobs.

**Organization**: Tasks are grouped by user story so each increment is independently implementable and testable.

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Capture the current CI performance baseline for comparison after optimization.

- [X] T001 Document pre-change CI durations and cache hit rates in `specs/001-optimize-ci-system-test/research.md`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Extend the reusable Go CI workflow so all downstream tasks can consume the shared artifact and metadata.

- [X] T002 **[Deprecated]** (Superseded by Phase 8 cache-only distribution) — The planned update to `github/devantler-tech/github-actions/reusable-workflows/.github/workflows/ci-go.yaml` to accept `artifact-name` and `artifact-checksum` inputs was not implemented; cache standardization replaced artifact-based distribution.
- [X] T003 **[Deprecated]** (Superseded by Phase 8 cache-only distribution) — The planned documentation update for artifact inputs in `github/devantler-tech/github-actions/reusable-workflows/README.md` was not implemented; see Phase 8 for cache usage instructions.

**Checkpoint**: Reusable workflow consumers can request the shared artifact via inputs. *(Deprecated: see Phase 8 for cache-based distribution.)*

---

## Phase 3: User Story 1 - Maintainer Receives Faster CI Feedback (Priority: P1) 🎯 MVP

**Goal**: Build the KSail binary once per workflow, reuse warmed Go caches, and cut system-test matrix runtimes below the 105-second target.

**Independent Test**: Trigger the workflow on a pull request with no code changes; verify `build-artifact` runs once, system-test jobs restore the cached binary (or rebuild on cache miss), and system-test matrix entries finish ≤105 seconds.

### Implementation for User Story 1

- [X] T004 [US1] Introduce `build-artifact` job in `.github/workflows/ci.yaml` that runs cached `go build`, records checksum, smoke-tests `./ksail --version`, and uploads `ksail-${{ github.run_id }}` outputs
- [X] T005 [P] [US1] Update the `pre-commit` job in `.github/workflows/ci.yaml` to standardize Go cache usage and remove redundant `go mod download` commands, without consuming the cached binary
- [X] T006 [P] [US1] Update the `ci` reusable-workflow invocation in `.github/workflows/ci.yaml` to rely on standardized action caching for Go modules without artifact passing, keeping lint and test jobs building from source
- [X] T007 [US1] Update the `system-test` matrix job in `.github/workflows/ci.yaml` to restore the cached binary per matrix entry, run the smoke step, and fall back to local `go build` on cache miss
- [X] T008 [US1] Update the `system-test-status` job in `.github/workflows/ci.yaml` to guard execution with `if: needs.build-artifact.result != 'success'` and short-circuit when the build job fails
- [X] T009 [US1] Standardize all Go jobs in `.github/workflows/ci.yaml` on `actions/setup-go@v6` with `cache-dependency-path: src/go.sum` and consistent cache keys
- [X] T010 [US1] Trigger `.github/workflows/ci.yaml` on a draft pull request and confirm every original job and matrix command still executes unchanged, recording findings in `specs/001-optimize-ci-system-test/research.md` (covers FR-007)

**Checkpoint**: Workflow builds once, system-test jobs reuse the cached binary, and Go module cache warming is enabled for all jobs.

---

## Phase 4: User Story 2 - Maintainer Diagnoses a Failing CI Job (Priority: P2)

**Goal**: Provide artifact lineage, cache status, and runtime metrics for every job so failures are attributable without rerunning builds.

**Independent Test**: Force a single system-test matrix failure and confirm the job summary lists artifact version, cache hit/miss, and duration while upstream jobs fail fast when the artifact is missing.

### Implementation for User Story 2

- [X] T011 [P] [US2] Add failing shell test `tests/scripts/collect-metrics_test.sh` that asserts metrics output includes duration, cache status, and artifact checksum before implementing the script
- [X] T012 [US2] Create job metrics helper script at `.github/scripts/collect-metrics.sh` to emit duration, cache status, and artifact checksum to `$GITHUB_STEP_SUMMARY`
- [X] T013 [US2] Add guard steps to `system-test` and `system-test-status` jobs in `.github/workflows/ci.yaml` that fail immediately when `needs.build-artifact.result != 'success'`
- [X] T014 [US2] Invoke `.github/scripts/collect-metrics.sh` within the `build-artifact` and `pre-commit` jobs in `.github/workflows/ci.yaml` to publish metrics and cache results
- [X] T015 [US2] Invoke `.github/scripts/collect-metrics.sh` within `system-test` and `system-test-status` jobs in `.github/workflows/ci.yaml`, capturing matrix durations and artifact metadata
- [X] T016 [US2] Extend `github/devantler-tech/github-actions/reusable-workflows/.github/workflows/ci-go.yaml` to call the metrics script (or equivalent shell) and surface cache-hit outputs for each reusable job
- [X] T017 [US2] Add a `metrics-summary` job in `.github/workflows/ci.yaml` that aggregates job outputs and appends SC-001–SC-005 data to `$GITHUB_STEP_SUMMARY`

**Checkpoint**: Every job emits diagnostics, guards block artifact-less runs, and the workflow summary aggregates performance data.

---

## Phase 5: User Story 3 - Contributor Introduces a New CI Job or Scenario (Priority: P3)

**Goal**: Make artifact reuse and caching opt-in by default so new jobs automatically adopt the optimized workflow without extra configuration.

**Independent Test**: Add a temporary matrix entry in a draft pull request and confirm it downloads the shared binary via the shared helper, uses warmed caches, and finishes within target runtime.

### Implementation for User Story 3

- [X] T018 [P] [US3] Add failing GitHub Actions workflow test `tests/actions/use-ksail-artifact.yml` that expects the composite action to download and verify the artifact
- [X] T019 [US3] Create composite action `.github/actions/use-ksail-artifact/action.yaml` to download the artifact, verify checksum, and run the smoke test based on supplied inputs
- [X] T020 **[Deprecated]** (Superseded by Phase 8 cache-only distribution) — The planned refactor of jobs in `.github/workflows/ci.yaml` to use `./.github/actions/use-ksail-artifact` was not implemented; cache standardization replaced artifact-based distribution.
- [X] T021 [US3] Update `github/devantler-tech/github-actions/reusable-workflows/.github/workflows/ci-go.yaml` to invoke the new composite action when artifact inputs are present so additional jobs require no manual wiring
- [X] T022 [US3] Update `specs/001-optimize-ci-system-test/quickstart.md` with instructions for using the composite action when adding new CI jobs or matrix entries

**Checkpoint**: Additional jobs inherit artifact reuse automatically and documentation explains the integration pattern.

---

## Final Phase: Polish & Cross-Cutting Concerns

**Purpose**: Finalize documentation, annotate workflow intent, and capture post-change benchmarks.

- [X] T023 Update inline comments in `.github/workflows/ci.yaml` describing artifact naming, cache keys, and guard expectations
- [X] T024 Record post-change workflow metrics in `specs/001-optimize-ci-system-test/research.md`, comparing against the baseline captured in T001 (covers SC-001–SC-004)
- [X] T025 Analyze non-artifact jobs (e.g., lint, pre-commit) to confirm runtime changes stay within the ≤10% regression threshold documented in `specs/001-optimize-ci-system-test/research.md` (covers SC-005)
- [X] T026 Review the first ten post-merge runs to confirm system-test pass rate matches the baseline and log findings in `specs/001-optimize-ci-system-test/research.md` (covers SC-006)
- [X] T027 Summarize optimization steps and metrics deltas in `specs/001-optimize-ci-system-test/quickstart.md`

---

## Phase 6: Metrics Rollback (Maintenance)

**Purpose**: Remove optional observability wiring introduced in US2 after maintainer feedback while preserving artifact reuse improvements.

- [X] T028 Retire metrics script outputs from `.github/workflows/ci.yaml`, removing summary jobs and references to `needs.*.outputs.metrics`
- [X] T029 Replace `.github/scripts/collect-metrics.sh` with a no-op placeholder and update `tests/scripts/collect-metrics_test.sh` to assert the deprecation message
- [X] T030 Refresh `specs/001-optimize-ci-system-test/{plan,quickstart,research,data-model}.md` to reflect the absence of metrics instrumentation

---

## Phase 7: Continuous Improvement - Cross-Run Artifact Cache

**Purpose**: Reuse the compiled `ksail` binary across workflow runs when source inputs are unchanged to shave the build job runtime.

- [X] T031 Add cache restore/save steps to the `build-artifact` job in `.github/workflows/ci.yaml` so cache hits skip recompilation while preserving smoke tests only (artifact uploads are removed in later steps)
- [X] T032 Update documentation (`specs/001-optimize-ci-system-test/research.md`, `quickstart.md`) to record the cache strategy and invalidation key components

---

## Phase 8: Cache-Only Distribution (Maintenance)

**Purpose**: Retire per-run artifacts and rely exclusively on the cache-backed binary.

> **Note:** This phase supersedes the shared artifact consumption approach originally introduced in Phase 2. The change from artifact-based distribution to cache-only distribution clarifies the historical context and evolution of the CI optimization strategy.

- [X] T033 Remove artifact upload/download steps from `.github/workflows/ci.yaml` and ensure system-test jobs restore or rebuild the binary from cache
- [X] T034 Update supporting documentation (plan, research, quickstart) to describe cache-only distribution and the deprecation of `.github/actions/use-ksail-artifact`

---

## Dependencies & Execution Order

- **Phase Order**: Setup → Foundational → User Story 1 → User Story 2 → User Story 3 → Polish.
- **Story Dependencies**: US1 is required before US2 and US3. US2 and US3 can proceed in parallel once US1 completes. Polish runs last after all targeted stories.
- **Intra-Story**:
  - US1: Complete T004 before starting T005–T010. T005 and T006 ([P]) can run concurrently once the build job exists.
  - US2: Complete T011 before T012–T017. T014 and T015 depend on the guard logic from T013.
  - US3: Complete T018 before refactors T019–T021; documentation T022 follows the refactor.

## Parallel Execution Examples

- **US1**: After T004 establishes the build job outputs, tasks T005 [P] and T006 [P] can be implemented simultaneously by different contributors (pre-commit vs. reusable CI wiring).
- **US2**: T011 [P] adds the failing test while T013 prepares guard conditions, expediting integration once both land.
- **US3**: T018 [P] creating the failing workflow test can run in parallel with preparatory review of `.github/workflows/ci.yaml`, enabling rapid refactors in T020 once the action is available.

## Implementation Strategy

1. **MVP (US1)**: Deliver artifact build job and consumer updates (T001–T010) so the workflow immediately benefits from faster runs.
2. **Diagnostics (US2)**: Layer on guardrails and metrics (T011–T017) to validate improvements and aid troubleshooting.
3. **Scalability (US3)**: Introduce the composite action and documentation (T018–T022) to keep future jobs fast by default.
4. **Polish**: Update comments and capture post-change metrics (T023–T027) before closing the feature.
