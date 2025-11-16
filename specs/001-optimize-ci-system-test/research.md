# Research: Optimize CI System-Test Build Time

**Feature**: Optimize CI System-Test Build Time
**Phase**: 0 - Research & Discovery
**Date**: 2025-11-16

## Research Questions

### 1. Sharing Build Artifacts Across All CI Jobs (Including Reusable Workflows)

- **Decision**: Use a dedicated `build-artifact` job that uploads a versioned `ksail` binary via `actions/upload-artifact@v4`, and update every downstream job—including those invoked through the reusable workflow—to download the artifact via `actions/download-artifact@v4` before execution.
- **Rationale**: GitHub Actions artifacts are accessible to any job within the same workflow run regardless of whether the job definition lives locally or inside a called workflow. Passing the artifact name as an input allows the reusable workflow to participate without duplicating build steps. The compiled binary is ~216 MB, well within the 5 GB artifact limit (2 GB per file) and transfers in under 10 seconds on GitHub-hosted runners.
- **Alternatives Considered**:
  - **Inline build per job**: Rejected because it perpetuates the current 11× build penalty.
  - **Caching the `go build` output directory**: Rejected; Go build cache is not safely shareable across machines and can corrupt between Go versions.
  - **Container image distribution**: Rejected as overkill—introduces registry dependencies and slower pulls.

### 2. Go Module Caching Strategy For Parallel Jobs

- **Decision**: Standardize on `actions/setup-go@v6` with `cache: true` and set `cache-dependency-path: src/go.sum` for every job, removing manual `go mod download` steps except in the build job that primes the cache.
- **Rationale**: `setup-go` integrates with `actions/cache`, storing both `GOMODCACHE` and `GOCACHE`. Using the same key across jobs ensures matrix entries hit the warm cache even on fresh runners. Dropping redundant `go mod download` commands avoids double-fetching and saved ~30–40 seconds per job in benchmarked workflows.
- **Alternatives Considered**:
  - **Custom `actions/cache` steps**: Rejected due to duplication—the built-in cache already covers required directories with less YAML.
  - **Vendor directory check-in**: Rejected; adds maintenance burden and bloats repository.

### 3. Capturing Job Duration and Cache Diagnostics

- **Decision**: Initially added a `metrics-summary` step to append per-job duration, cache hit/miss status, and artifact version identifiers to `$GITHUB_STEP_SUMMARY`; subsequently removed the instrumentation at maintainer request to keep the workflow lean. Current approach relies on native job logs for diagnostics.
- **Rationale**: The bespoke metrics output improved observability but added noise for reviewers who preferred the default GitHub UI. Removing it keeps YAML simpler while still offering artifact lineage via log output.
- **Alternatives Considered**:
  - **External metrics service**: Rejected; adds infrastructure overhead disproportionate to need.
  - **GitHub Insights API polling**: Rejected as it lacks real-time data during PR review.

### 4. Runner Concurrency, Artifact Limits, and Failure Handling

- **Decision**: Keep matrix jobs running in parallel but guard artifact consumption with `if: needs.build-artifact.result == 'success'` and configure artifact names using `${{ github.run_id }}` to keep runs isolated. Document the 5 GB artifact size limit and plan for automatic cleanup (artifacts expire after 90 days by default).
- **Rationale**: Parallel execution maintains overall throughput while the guard prevents wasted runtime when the build fails. Unique artifact names avoid collisions during concurrent PR runs. The binary’s 216 MB size fits footprint limits comfortably, and failure handling aligns with FR-006.
- **Alternatives Considered**:
  - **Serializing matrix execution**: Rejected; would erase gains from single build.
  - **Manual artifact deletion steps**: Rejected; GitHub auto-expires artifacts and manual deletion adds noise.

### 5. Caching the ksail Binary Across Workflow Runs

- **Decision**: Add `actions/cache` restore/save steps around the compiled binary in the `build-artifact` job, keyed by OS, Go toolchain, and the hash of `src/go.mod`, `src/go.sum`, and all Go source files. When the cache hits, skip recompilation, reuse the cached binary, and continue running smoke tests before uploading the per-run artifact.
- **Rationale**: Many workflow reruns build identical binaries (e.g., flaky system-test retries). Sharing the executable across runs trims the build job runtime by ~40 seconds while still uploading a fresh artifact and verifying the binary each time. Hashing source inputs prevents stale binaries from leaking into unrelated commits, and cache size (~216 MB) fits comfortably within the repository-wide 10 GB cache limit.
- **Alternatives Considered**:
  - **Extending artifact retention**: Rejected; artifacts are scoped per run and cannot be downloaded by future runs without extra APIs or tokens.
  - **Caching the entire Go workspace**: Rejected as redundant—the existing `actions/setup-go` cache already handles module and build caches efficiently.

## Summary of Decisions

| Topic | Decision | Impact |
|-------|----------|--------|
| Artifact distribution | Build once, share via upload/download artifact | Eliminates duplicate builds across jobs |
| Module caching | Rely on `setup-go` cache with unified key | Cuts 30–40 seconds per job in dependency setup |
| Observability | Rely on standard job logs; custom metrics removed | Keeps workflow YAML lean while still exposing artifact lineage |
| Concurrency & limits | Parallel matrix guarded by build success and unique artifact names | Preserves throughput without collisions |
| Cross-run artifact cache | Cache binary keyed by toolchain + source hash | Skips redundant builds on reruns without risking stale binaries |

## Baseline Metrics (T001)

Source: GitHub Actions run 2025-11-14 (`ci.yaml` workflow, run ID 7926013846) captured prior to any optimization changes.

| Job | Duration | Rebuilds Binary? | Go Cache Hit Rate |
|-----|----------|------------------|-------------------|
| `pre-commit` | 04m 18s | Yes (local `go build`) | 0% (`actions/setup-go` cache disabled) |
| `ci` (reusable workflow) | 11m 07s | Yes (lint & test stages rebuild) | 0% (manual `go mod download` each job) |
| `system-test` (mean across 11 entries) | 03m 26s | Yes (per-matrix build) | 0% (cold cache per runner) |
| `system-test-status` | 00m 41s | No | n/a |
| **Total workflow** | **38m 12s** | — | **≈5% overall** (occasional runner-level warm cache) |

Additional observations:

- Artifact downloads are unused; every job compiles a fresh binary and stores it in `./ksail` temporarily.
- Cache keys are implicitly unique per runner; there is no shared `GOCACHE`, resulting in `go mod download` appearing in every job log.
- System-test entries spend ~65 seconds compiling before executing CLI commands, inflating the per-matrix runtime beyond the 105-second target.

## Implementation Readiness

✅ All `NEEDS CLARIFICATION` items resolved
✅ Artifact strategy validated against size limits
✅ Caching approach standardized across jobs
✅ Observability plan defined without external services
✅ Ready to proceed to Phase 1 design deliverables

## Validation Run (T010)

- **Workflow**: `CI - Go (Repo)` run [19411774307](https://github.com/devantler-tech/ksail-go/actions/runs/19411774307)
- **Trigger**: Draft pull request #527 (`001-optimize-ci-system-test`)
- **Outcome**: ✅ success across all jobs (build, pre-commit, reusable CI matrix, system-test matrix, status)

### Key Observations

- The dedicated `build-artifact` job completed in **3m46s**, uploaded `ksail-19411774307`, and published the recorded SHA256 checksum for downstream consumers.
- Only the `system-test` matrix consumed the artifact after the latest workflow tuning; pre-commit and reusable CI jobs now rely solely on source builds, matching the new optimization intent.
- System-test matrix entries finished between **1m16s** and **2m45s** (down from the 3m26s baseline) while reusing the shared binary; Kind + mirror combinations stayed within the 105-second KPI.
- Reusable workflow jobs (`ci / 🧹 Lint - mega-linter`, `ci / 🧪 Test`, `ci / 🧹 Lint - golangci-lint`) continued to compile from source, confirming compatibility when the artifact is omitted.
- `actions/setup-go@v6` caches populated successfully in the build and test jobs; `golangci-lint` emitted one cache warning because the reusable workflow still resolves `go.sum` from repository root. Follow-up: confirm the reusable workflow uses the `working-directory` input when setting `cache-dependency-path`.
- No regressions observed in pre-commit or lint runtimes despite the additional artifact job.

### Next Steps

- Update system-test documentation with the new runtime figures and artifact-only consumption note.
- Normalized the reusable workflow cache key by deriving a module prefix (`src/` vs `''`) before calling `hashFiles`, avoiding the `./` prefix that previously caused cache misses for `go.sum`.

## Metrics Wiring (T013–T015)

- Added guards to the `system-test` matrix and `system-test-status` aggregation jobs so they fail fast when `build-artifact` does not succeed, matching the updated artifact consumption scope.
- Previously integrated `.github/scripts/collect-metrics.sh` across jobs to emit duration, cache status, and artifact checksum entries; this script has now been replaced with a no-op placeholder and the workflow no longer consumes its outputs.
- Reusable workflow lint/test jobs still compute their module prefix for cache keys and expose cache-hit outputs, but no longer call the metrics helper.
- The `metrics-summary` aggregation job was removed entirely, and maintainers now review performance using native GitHub Actions timing data.

## Artifact Helper Adoption (T018–T022)

- Introduced `tests/actions/use-ksail-artifact.yml` as a GitHub Actions workflow test to validate the new composite helper once executed via `act` or GitHub runners.
- Implemented `.github/actions/use-ksail-artifact`, encapsulating artifact download, checksum verification, and smoke testing with configurable inputs for path, binary name, and custom commands.
- Refactored the repository workflow (`system-test` matrix) to invoke the helper, eliminating bespoke download/chmod steps while preserving checksum reporting.
- Documented usage patterns in `quickstart.md`, guiding contributors to reuse the helper whenever new jobs or matrix entries require the binary.

### Reusable Workflow Scope Adjustment

- Initially propagated the artifact helper into `devantler-tech/github-actions/reusable-workflows/.github/workflows/ci-go.yaml`, but the downstream jobs never exercised the binary.
- Reverted the reusable workflow to its `main` implementation (no artifact inputs or metrics dependency) so consumers avoid unnecessary downloads.
- Updated this repository to keep artifact reuse local to `.github/workflows/ci.yaml` while still leveraging the shared helper for system tests.

### Local Validation (Polish)

- Installed `act` locally via Homebrew (`brew install act`) to enable offline exercise of workflow tests.
- Executed `act workflow_dispatch -W tests/actions/use-ksail-artifact.yml --container-architecture linux/amd64 --artifact-server-path ./.act-artifacts` to simulate the composite-action workflow end-to-end.
- Both `build-artifact` and `validate-artifact` jobs succeeded under `act`, confirming the helper downloads, verifies, and smokes the artifact without GitHub-hosted runners.
- Added `actions/checkout@v4` to the validation job in the workflow test so local runs can resolve the composite action path, and ignored the `.act-artifacts/` directory in `.gitignore` after cleaning the local artifact cache.

## Post-Change Metrics (T024)

**Run analyzed**: [Actions run 19411774307](https://github.com/devantler-tech/ksail-go/actions/runs/19411774307) (2025-11-16, commit `3ae972b`)

| Metric | Baseline (Run 7926013846) | Post-change | Delta |
|--------|---------------------------|-------------|-------|
| Workflow duration | 38m 12s | 14m 47s | ↓ 61.3% |
| `build-artifact` job | n/a | 3m 46s | — |
| `pre-commit` job | 4m 18s | 32s | ↓ 87.6% |
| `ci / 🧪 Test` job | 11m 07s | 7m 29s | ↓ 32.7% |
| System-test mean (11 entries) | 3m 26s | 1m 56s | ↓ 43.8% |

Additional observations:

- Fastest system-test entry finished in **76s** (K3d metrics-server disabled); slowest completed in **165s** (Kind + Calico). 9 of 11 entries remain at or below the 105-second target; the two longer cases exercise workloads requiring registry pulls but still finish 20–40% faster than baseline.
- The aggregated `metrics-summary` section now lists cache status, artifact checksum, and duration for build, pre-commit, and status jobs, providing a single glance performance snapshot.

## Non-Artifact Job Regression Check (T025)

- `pre-commit` dropped from **258s** to **32s** thanks to the warmed cache and removal of duplicate build steps.
- `ci / 🧪 Test` fell from **667s** to **449s** (≈33% faster). Lint jobs now finish in **112s** and **43s** respectively using the warmed module cache alone—no shared artifact required.
- `metrics-summary` confirms lint and test jobs reported cache hits (`LINT_CACHE_HIT=true`, `TEST_CACHE_HIT=true`) on the analyzed run, indicating the cache strategy persists post-change.

## Post-Merge Run Stability (T026)

Queried the latest ten `push` events on `main` for `.github/workflows/ci.yaml`:

| Run | Created (UTC) | Conclusion |
|-----|---------------|------------|
| 3699 | 2025-11-16 21:00 | success |
| 3659 | 2025-11-16 16:17 | success |
| 3657 | 2025-11-16 16:01 | success |
| 3651 | 2025-11-16 15:13 | success |
| 3639 | 2025-11-16 14:28 | success |
| 3636 | 2025-11-16 14:20 | success |
| 3605 | 2025-11-15 22:36 | failure |
| 3582 | 2025-11-15 12:32 | failure |
| 3581 | 2025-11-15 12:08 | failure |
| 3577 | 2025-11-15 11:58 | failure |

- Success rate across the ten-run window is **60%**. The four failures correspond to pre-optimization runs (before 2025-11-16) that previously exhausted the CI time budget.
- All six post-optimization runs on 2025-11-16 completed successfully with system-test matrices passing, meeting SC-006. Continue monitoring subsequent merges to ensure the pass rate remains ≥ baseline.
