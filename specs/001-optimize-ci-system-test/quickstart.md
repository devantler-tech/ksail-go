# Quickstart Guide: Optimize CI System-Test Build Time

**Feature**: Optimize CI System-Test Build Time
**Date**: 2025-11-16
**Audience**: Maintainers updating `.github/workflows/ci.yaml`

## Prerequisites

- GitHub CLI (`gh`) configured for the repository (optional but recommended)
- Familiarity with GitHub Actions syntax
- Ability to update both this repository and the `devantler-tech/reusable-workflows` repository if parameter changes are required
- Clean git working tree (`git status` shows no changes)

## Implementation Steps

### 1. Baseline Measurements

1. Open the latest successful run of `.github/workflows/ci.yaml` and note:
   - Average duration of `system-test` matrix jobs (should be ~206 seconds currently)
   - Total workflow duration (~38 minutes)
2. Record metrics in issue #522 for before/after comparison.

### 2. Add Dedicated Build Job

1. In `.github/workflows/ci.yaml`, add a new `build-artifact` job before `pre-commit` that:

   - Checks out code with `actions/checkout@v5`
   - Sets up Go via `actions/setup-go@v6` using `cache: true` and `cache-dependency-path: src/go.sum`
   - Runs `go build -C src -o ksail`
   - Executes `./ksail --version` (smoke test)
   - Computes SHA256 (`shasum -a 256 ksail`)
   - Uploads the binary with `actions/upload-artifact@v4` (name: `ksail-${{ github.run_id }}`)
   - Exposes `artifact-name` and `checksum` via job outputs
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

### 4. Update Reusable CI Workflow Consumption

1. In `devantler-tech/reusable-workflows`:
   - Add optional inputs `artifact-name` and `artifact-path` to `ci-go.yaml`.
   - Insert download + smoke steps in lint and test jobs when inputs are provided.
   - Remove standalone `go build -v ./...` if the binary is not required; rely on tests to compile packages and the artifact for CLI invocations.
2. In this repository’s `ci` job call (within `.github/workflows/ci.yaml`), pass the outputs from `build-artifact`:

   ```yaml
   with:
     working-directory: ./src/
     artifact-name: ${{ needs.build-artifact.outputs.artifact-name }}
     artifact-checksum: ${{ needs.build-artifact.outputs.checksum }}
   ```

3. Regenerate the reusable workflow tag or reference the new commit.

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
   - Ensure `build-artifact` runs once
   - Confirm system-test jobs complete in ≤105 seconds
   - Verify job summaries show cache and artifact data

### 9. Document Performance Delta

1. Update issue #522 with before/after numbers (average system-test duration, total workflow time, cache hit rate).
2. Note any jobs exceeding the 10 % regression threshold.

### 10. Post-Merge Follow-Up

1. Re-run `system-test` on `main` to confirm improvements persist.
2. Remove any temporary debugging outputs added during implementation.
3. Schedule a follow-up task if reusable workflow consumers in other repos need to adopt the new artifact inputs.

## Rollback Plan

- If artifact download fails across multiple jobs, revert to previous workflow commit and re-open issue with investigation notes.
- Keep the old YAML snippet in branch history to simplify rollback via `git revert`.

## Success Indicators

✅ `build-artifact` job runs once per workflow and publishes outputs

✅ System-test matrix jobs finish in ≤105 seconds with >80 % cache hit rate

✅ Total workflow duration ≤25 minutes

✅ Job summaries include duration, cache status, and artifact checksum

✅ No downstream job regresses by more than 10 %
