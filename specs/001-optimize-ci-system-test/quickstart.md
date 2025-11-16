# Quickstart Guide: Optimize CI System-Test Build Time

**Feature**: Optimize CI System-Test Build Time
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
- System-test matrix entries average **1m 56s** with the shared artifact; the slowest case (Kind + Calico) still improved 20% versus baseline.
- `metrics-summary` surfaces cache status and artifact checksums for every job, and first-six post-merge runs on `main` completed successfully.

## Implementation Steps

### 1. Monitor Current Benchmarks

Track the metrics emitted by the `metrics-summary` job using the following targets:

- Workflow duration: ~15 minutes
- System-test entries: ≤105 seconds for lightweight combinations, ≤165 seconds for workload-heavy scenarios
- Cache hit indicators: `LINT_CACHE_HIT=true`, `TEST_CACHE_HIT=true`
- Artifact checksum consistency between build and consumer jobs

### 2. Add Dedicated Build Job

1. In `.github/workflows/ci.yaml`, add a new `build-artifact` job before `pre-commit` that:

   - Checks out code with `actions/checkout@v5`
   - Set up Go via `actions/setup-go@v6` using `cache: true` and `cache-dependency-path: src/go.sum`
   - Run `go build -C src -o ksail`
   - Execute `./ksail --version` (smoke test)
   - Compute SHA256 (`shasum -a 256 ksail`)
   - Upload the binary with `actions/upload-artifact@v4` (name: `ksail-${{ github.run_id }}`)
   - Expose `artifact-name` and `checksum` via job outputs
2. Set `needs: [build-artifact]` on every downstream job.

### 3. Update Pre-Commit Job

1. Remove redundant `go mod download` step (cache provides modules).
2. Insert artifact download and smoke test before any hook invocation:

   ```yaml
   - uses: actions/download-artifact@v4
     with:
       name: ${{ needs.build-artifact.outputs.artifact-name }}
       path: ./bin
   - run: ./bin/ksail --version
   ```

3. Ensure `pre-commit` continues to call `pre-commit/action@v3.0.1` as last step.

### 4. Keep Reusable Workflow Lean

The shared workflow `devantler-tech/github-actions/reusable-workflows/.github/workflows/ci-go.yaml` remains unchanged—lint and test jobs build from source and rely solely on warmed module caches. No artifact inputs are required; all binary reuse happens inside this repository’s workflow via the composite helper.

If a future consumer genuinely needs the shared artifact, prefer adding helper steps directly in that repository instead of expanding the reusable workflow contract.

### 5. Optimize System-Test Job

1. Replace local build steps with artifact download + smoke test.
2. Ensure each matrix entry sets up Go with caching but skips `go mod download`.
3. Prepend each test command block (`cluster init`, `create`, etc.) with `./bin/ksail` path.
4. Capture job metrics:

   ```yaml
   - name: Record metrics
     run: |
       end=$SECONDS
       duration=$((end - env.START_TIME))
       echo "- Duration: ${duration}s" >> $GITHUB_STEP_SUMMARY
       echo "- Cache: ${{ steps.go.outputs.cache-hit }}" >> $GITHUB_STEP_SUMMARY
       echo "- Artifact SHA: ${{ needs.build-artifact.outputs.checksum }}" >> $GITHUB_STEP_SUMMARY
     env:
       START_TIME: ${{ steps.start.outputs.seconds }}
   ```

### 6. Add Workflow-Wide Summary

1. After `system-test-status`, append a `metrics-summary` job that:
   - Runs `if: always()`
   - Aggregates duration data from job outputs (use `fromJSON(needs.*.outputs.metrics)`)
   - Writes consolidated totals to `$GITHUB_STEP_SUMMARY`

2. The `ci` job continues to call the reusable workflow with only the `working-directory` input—artifact handling is intentionally scoped to this repository.

### 6.5 Use the Helper for New Jobs

When adding a new matrix entry or standalone job that requires the compiled binary, include a step similar to:

```yaml
- name: Prepare ksail binary
  uses: ./.github/actions/use-ksail-artifact
  with:
    artifact-name: ${{ needs.build-artifact.outputs.artifact-name }}
    artifact-checksum: ${{ needs.build-artifact.outputs.checksum }}
```

This guarantees artifact reuse, checksum validation, and smoke testing without duplicating YAML.

### 7. Validate Locally (Optional)


1. Use [`act`](https://github.com/nektos/act) to dry-run a reduced matrix (e.g., Kind default) verifying artifact download and smoke test succeed:

   ```bash
   act pull_request --job system-test --matrix init-args='--distribution Kind'
   ```

2. Confirm the binary step executes and `./bin/ksail --version` passes.

### 8. Push Branch and Observe CI

1. Commit changes (`git add .github` and supporting files).
2. Push to `001-optimize-ci-system-test`.
3. Monitor the workflow:
   - Ensure `build-artifact` runs once and reports the checksum recorded by downstream jobs
   - Confirm system-test jobs stay within the targets listed in **Monitor Current Benchmarks**
   - Review `metrics-summary` for cache hits and duration deltas

### 9. Document Performance Delta

1. Update issue #522 with before/after numbers (average system-test duration, total workflow time, cache hit rate).
2. Note any jobs exceeding the 10% regression threshold.

### 10. Post-Merge Follow-Up

1. Review the latest ten push runs on `main` via `gh run list --workflow "CI - Go (Repo)" --branch main --event push --limit 10`.
2. Confirm system-test pass rate remains at or above the baseline (currently 6/6 successes after the optimization landed).
3. If another repository needs artifact reuse, open a task to add the helper locally in that codebase rather than extending the shared workflow.

## Rollback Plan

- If artifact download fails across multiple jobs, revert to previous workflow commit and re-open issue with investigation notes.
- Keep the old YAML snippet in branch history to simplify rollback via `git revert`.

## Success Indicators

✅ `build-artifact` job runs once per workflow and publishes outputs

✅ System-test matrix jobs finish in ≤105 seconds with >80% cache hit rate

✅ Total workflow duration ≤25 minutes

✅ Job summaries include duration, cache status, and artifact checksum

✅ No downstream job regresses by more than 10%
