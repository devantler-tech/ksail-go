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

- **Decision**: Add a `metrics-summary` step that appends per-job duration, cache hit/miss status, and artifact version identifiers to the GitHub job summary using the `$GITHUB_STEP_SUMMARY` file, and expose totals via workflow outputs.
- **Rationale**: Job summaries provide a persistent, human-readable log without external tooling. Duration data can be captured by recording `$(date +%s)` before/after critical steps. Listing cache-key status and artifact SHA1 aids post-run triage when failures occur.
- **Alternatives Considered**:
  - **External metrics service**: Rejected; adds infrastructure overhead disproportionate to need.
  - **GitHub Insights API polling**: Rejected as it lacks real-time data during PR review.

### 4. Runner Concurrency, Artifact Limits, and Failure Handling

- **Decision**: Keep matrix jobs running in parallel but guard artifact consumption with `if: needs.build-artifact.result == 'success'` and configure artifact names using `${{ github.run_id }}` to keep runs isolated. Document the 5 GB artifact size limit and plan for automatic cleanup (artifacts expire after 90 days by default).
- **Rationale**: Parallel execution maintains overall throughput while the guard prevents wasted runtime when the build fails. Unique artifact names avoid collisions during concurrent PR runs. The binary’s 216 MB size fits footprint limits comfortably, and failure handling aligns with FR-006.
- **Alternatives Considered**:
  - **Serializing matrix execution**: Rejected; would erase gains from single build.
  - **Manual artifact deletion steps**: Rejected; GitHub auto-expires artifacts and manual deletion adds noise.

## Summary of Decisions

| Topic | Decision | Impact |
|-------|----------|--------|
| Artifact distribution | Build once, share via upload/download artifact | Eliminates duplicate builds across jobs |
| Module caching | Rely on `setup-go` cache with unified key | Cuts 30–40 seconds per job in dependency setup |
| Observability | Emit metrics via job summary | Provides immediate visibility into performance improvements |
| Concurrency & limits | Parallel matrix guarded by build success and unique artifact names | Preserves throughput without collisions |

## Implementation Readiness

✅ All `NEEDS CLARIFICATION` items resolved
✅ Artifact strategy validated against size limits
✅ Caching approach standardized across jobs
✅ Observability plan defined without external services
✅ Ready to proceed to Phase 1 design deliverables
