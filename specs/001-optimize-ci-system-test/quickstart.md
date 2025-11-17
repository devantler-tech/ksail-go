# Quickstart Guide: Optimize CI System Test Build Time

**Feature**: Optimize CI System Test Build Time
**Date**: 2025-11-16
**Audience**: Maintainers updating `.github/workflows/ci.yaml`

## Prerequisites

- GitHub CLI (`gh`) configured for the repository (optional but recommended)
- Familiarity with GitHub Actions syntax
- Ability to coordinate with `devantler-tech/reusable-workflows` maintainers if future parameter changes become necessary
- Clean git working tree (`git status` shows no changes)

## Key Outcomes (Post-Change Metrics)

- Total workflow runtime fell from **38m 12s** to **14m 47s** on run [19411774307](https://github.com/devantler-tech/ksail-go/actions/runs/19411774307).
- `pre-commit` now finishes in **32s** (down 88%); the reusable workflow’s longest job (`ci / 🧪 Test`) completes in **7m 29s** (down 33%).
- System-test matrix entries average **1m 56s** when the cache seeds the shared binary; the slowest case (Kind + Calico) still improved 20% versus baseline.
- Even without custom metrics instrumentation, the first six post-merge runs on `main` completed successfully and noted cache hit status alongside the binary checksum in standard job logs.

## Implementation Steps

### 1. Monitor Current Benchmarks

Record baseline numbers directly from the GitHub Actions run details page. Focus on:

- Workflow duration: ~15 minutes target after optimizations
- System-test entries: ≤105 seconds for lightweight combinations, ≤165 seconds for workload-heavy scenarios
- Cache hit indicators reported by `actions/setup-go`
- Binary smoke test output remains consistent between build and consumer jobs (visible in job logs)

### 2. Add Dedicated Build Job

1. In `.github/workflows/ci.yaml`, add a new `build-artifact` job before `pre-commit` that:

   - Check out code with `actions/checkout@v5`
   - Set up Go via `actions/setup-go@v6` using `cache: true` and `cache-dependency-path: src/go.sum`
   - Restore the cached `ksail` binary with `actions/cache@v4`, keyed by OS, Go version, and a hash of `src/go.mod`, `src/go.sum`, and all Go source files; when the cache hits, skip recompilation but still run the smoke test
   - Run `go build -C src -o ../ksail .` when the cache misses, seed `.cache/ksail`, and save the cache for future runs
   - Execute `./ksail --version` (smoke test)
2. Keep downstream jobs dependent on `build-artifact` so they only start after the cache is populated (or the fallback build completes).

### 3. Update Pre-Commit Job

1. Remove redundant `go mod download` step (the Go cache provides modules).
2. Keep `pre-commit` focused on linting/formatting; it does not need the compiled binary once cache sharing is in place.
3. Ensure `pre-commit` continues to call `pre-commit/action@v3.0.1` as the final step.

### 4. Keep Reusable Workflow Lean

The shared workflow `devantler-tech/github-actions/reusable-workflows/.github/workflows/ci-go.yaml` remains unchanged—lint and test jobs build from source and rely solely on warmed module caches. No additional inputs are required; binary reuse is handled locally via cache restores.

### 5. Optimize System Test Job

1. Restore the cached `ksail` binary using the same cache key as the build job.
2. If the cache misses, run `go build -C src -o ../bin/ksail .` to produce the binary locally, copy it into `.cache/ksail`, and save the cache for subsequent runs.
3. Execute `./bin/ksail --version` before the matrix commands to maintain the smoke guard, then run the usual suite (`cluster init`, `create`, etc.).

### 6. Reuse the Cached Binary in New Jobs

When adding a new job or matrix entry that needs the compiled binary, restore the cache with the same key formula used in the build and system-test jobs. Ensure the step builds locally on cache miss, saves the cache, and runs a quick smoke command before using the binary.

### 7. Validate Locally (Optional)

> **Note**: `act` does not support GitHub Actions cache restoration by default. The validation below is limited to smoke tests only and will not verify cache behavior. For comprehensive cache validation, rely on the CI workflow runs described in step 8.

1. Use [`act`](https://github.com/nektos/act) to dry-run a reduced matrix (e.g., Kind default) verifying the fallback build and smoke test succeed:

   ```bash
   act pull_request --job system-test --matrix init-args='--distribution Kind'
   ```

2. Confirm the binary step executes (via fallback build) and `./bin/ksail --version` passes.

**Alternative**: To validate cache behavior, push to a feature branch and review the GitHub Actions run logs for:

- Cache hit/miss indicators in the `build-artifact` job output
- Binary reuse confirmation in system-test job logs
- Overall timing improvements compared to baseline runs

### 8. Push Branch and Observe CI

1. Commit changes (`git add .github` and supporting files).
2. Push to `001-optimize-ci-system-test`.
3. Monitor the workflow:
   - Ensure `build-artifact` runs once, logs whether the binary came from cache, and reseeds the cache when rebuilding
   - Confirm system-test jobs stay within the targets listed in **Monitor Current Benchmarks**
   - Use the GitHub Actions job pages to confirm cache hits and duration deltas where available

### 9. Document Performance Delta

1. Update issue #522 with before/after numbers (average system-test duration, total workflow time, cache hit rate).
2. Note any jobs exceeding the 10% regression threshold.

### 10. Post-Merge Follow-Up

1. Review the latest ten push runs on `main` via `gh run list --workflow "CI - Go (Repo)" --branch main --event push --limit 10`.
2. Confirm system-test pass rate remains at or above the baseline (currently 6/6 successes after the optimization landed).
3. If another repository needs to reuse the compiled binary, replicate this cache-restore pattern within that codebase instead of reintroducing shared artifacts.

## Rollback Plan

- If cache restores fail across multiple jobs (forcing repeated fallback builds), revert to the previous workflow commit and re-open the issue with investigation notes.
- Keep the old YAML snippet in branch history to simplify rollback via `git revert`.

## Success Indicators

✅ `build-artifact` job runs once per workflow and seeds the binary cache

✅ System-test matrix jobs finish in ≤105 seconds with >80% cache hit rate

✅ Total workflow duration ≤25 minutes

✅ No downstream job regresses by more than 10%
