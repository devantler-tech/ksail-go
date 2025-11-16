# Workflow Contracts

**Feature**: Optimize CI System-Test Build Time
**Phase**: 1 - Design
**Date**: 2025-11-16

## Overview

These contracts define the mandatory behaviors for each CI job after the optimization. They ensure the single build artifact, shared caches, and observability hooks function consistently across the workflow.

## Build Artifact Contract

**Producer**: `build-artifact` job

```yaml
- name: "Compile ksail binary"
  run: "go build -C src -o ../ksail"
  outputs:
    artifact-name: "ksail-${{ github.run_id }}"
    checksum: "${{ steps.hash.outputs.sha256 }}"
  assertions:
    - file_exists: "ksail"
    - file_size_lt: "2147483648"   # 2 GB per-file limit
    - exit_code: 0

- name: "Smoke test binary"
  run: "./ksail --version"
  assertions:
    - exit_code: 0

- name: "Upload artifact"
  uses: actions/upload-artifact@v4
  with:
    name: "ksail-${{ github.run_id }}"
    path: "ksail"
  assertions:
    - upload_result: "success"
```

**Success Criteria**: Artifact uploaded with checksum published and accessible to downstream jobs.

## Artifact Consumption Contract

**Consumers**: `pre-commit`, `ci` (lint, mega-lint, test via reusable workflow), `system-test`, `system-test-status`

```yaml
- name: "Download ksail binary"
  uses: actions/download-artifact@v4
  with:
    name: "ksail-${{ needs.build-artifact.outputs.artifact-name }}"
    path: ./bin
  conditions:
    - needs.build-artifact.result == 'success'
  assertions:
    - directory_exists: "./bin"
    - file_exists: "./bin/ksail"

- name: "Validate artifact"
  run: "./bin/ksail --version"
  assertions:
    - exit_code: 0
    - stdout_contains: ${{ github.sha }}
```

**Failure Behavior**: If download or smoke test fails, job must exit with non-zero status, log message, and skip subsequent steps.

## Go Module Cache Contract

```yaml
- name: "Setup Go"
  uses: actions/setup-go@v6
  with:
    go-version-file: src/go.mod
    cache: true
    cache-dependency-path: src/go.sum
  outputs:
    cache-hit: ${{ steps.go-cache.outputs.cache-hit }}

- name: "Record cache status"
  run: |
    if [ "${{ steps.go-cache.outputs.cache-hit }}" = "true" ]; then
      echo "cache=hit" >> $GITHUB_STEP_SUMMARY
    else
      echo "cache=miss" >> $GITHUB_STEP_SUMMARY
    fi
```

**Success Criteria**: Every job reports cache hit/miss and completes without redundant `go mod download` steps (except build job).

## Metrics Contract

```yaml
- name: "Capture job metrics"
  run: |
    start=$SECONDS
    # ... job steps ...
    duration=$((SECONDS - start))
    {
      echo "### Job Metrics"
      echo "- Duration: ${duration}s"
      echo "- Cache: $CACHE_STATUS"
      echo "- Artifact SHA256: $ARTIFACT_SHA"
    } >> "$GITHUB_STEP_SUMMARY"
  env:
    CACHE_STATUS: ${{ steps.go-cache.outputs.cache-hit && 'hit' || 'miss' }}
    ARTIFACT_SHA: ${{ needs.build-artifact.outputs.checksum || 'n/a' }}
```

**Success Criteria**: Summary includes duration, cache status, and artifact checksum for every job; absence is treated as regression.

## Guard Contract

```yaml
- name: "Ensure build succeeded"
  if: needs.build-artifact.result != 'success'
  run: |
    echo "Build artifact unavailable. Skipping job." >> $GITHUB_STEP_SUMMARY
    exit 1
```

**Purpose**: Prevents downstream jobs from running when the build fails; fulfills FR-006.

## Post-Run Aggregation Contract

```yaml
- name: "Publish workflow summary"
  if: always()
  run: |
    echo "### CI Performance Snapshot" >> $GITHUB_STEP_SUMMARY
    echo "- Total duration: ${{ steps.aggregate.outputs.total-seconds }}s" >> $GITHUB_STEP_SUMMARY
    echo "- System-test avg: ${{ steps.aggregate.outputs.system-test-avg }}s" >> $GITHUB_STEP_SUMMARY
    echo "- Cache hit rate: ${{ steps.aggregate.outputs.cache-hit-rate }}" >> $GITHUB_STEP_SUMMARY
```

**Success Criteria**: Maintainers receive a consolidated snapshot aligning with SC-001–SC-005.

## Notes

- Contracts apply equally to jobs defined locally and those invoked via reusable workflows; new inputs/outputs may be required in the reusable workflow to honor them.
- Any deviation (missing artifact, cache misconfiguration, absent metrics) must fail the job and surface in the summary for fast remediation.
