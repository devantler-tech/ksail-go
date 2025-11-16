# Workflow Contracts

**Feature**: Optimize CI System-Test Build Time
**Phase**: 1 - Design
**Date**: 2025-11-16

## Overview

These contracts define the mandatory behaviors for each CI job after the optimization. They ensure the cached binary, shared Go module caches, and guard mechanisms function consistently across the workflow.

## Build Binary Cache Contract

**Producer**: `build-artifact` job

```yaml
- name: "Compute ksail cache key"
  id: ksail-cache-key
  env:
    KSAIL_CACHE_KEY: ${{ format('{0}-ksail-bin-{1}-{2}', runner.os, steps.setup-go.outputs.go-version, hashFiles('src/go.mod', 'src/go.sum', 'src/**/*.go')) }}
  run: |
    printf 'value=%s\n' "$KSAIL_CACHE_KEY" >> "$GITHUB_OUTPUT"

- name: "Restore cached ksail binary"
  id: ksail-cache
  uses: actions/cache/restore@v4
  with:
    path: ./.cache/ksail
    key: ${{ steps.ksail-cache-key.outputs.value }}

- name: "Use cached binary"
  if: steps.ksail-cache.outputs.cache-hit == 'true'
  run: cp ./.cache/ksail ./ksail

- name: "Build ksail binary"
  if: steps.ksail-cache.outputs.cache-hit != 'true'
  run: go build -C src -o ../ksail .
  assertions:
    - file_exists: "ksail"
    - exit_code: 0

- name: "Store binary for cache"
  if: steps.ksail-cache.outputs.cache-hit != 'true'
  run: |
    mkdir -p ./.cache
    cp ./ksail ./.cache/ksail

- name: "Save ksail binary cache"
  if: steps.ksail-cache.outputs.cache-hit != 'true'
  uses: actions/cache/save@v4
  with:
    path: ./.cache/ksail
    key: ${{ steps.ksail-cache-key.outputs.value }}

- name: "Smoke test"
  run: "./ksail --version"
  assertions:
    - exit_code: 0
```

**Success Criteria**: Binary cached with deterministic key and accessible to downstream jobs via cache restore.

## Binary Cache Consumption Contract

**Consumers**: `system-test`

> **Note:** As of Phase 8 (T033-T034), artifact distribution is deprecated. The `system-test` job consumes the built `ksail` binary via cache restore. The `pre-commit` job and reusable CI workflow jobs (`ci`) build from source and do not consume the cached binary.

```yaml
- name: "Verify build artifact"
  if: needs.build-artifact.result != 'success'
  run: |
    echo "build-artifact failed to seed the ksail binary cache. Failing system-test matrix." >> "$GITHUB_STEP_SUMMARY"
    exit 1

- name: "Compute ksail cache key"
  id: ksail-cache-key
  env:
    KSAIL_CACHE_KEY: ${{ format('{0}-ksail-bin-{1}-{2}', runner.os, steps.setup-go.outputs.go-version, hashFiles('src/go.mod', 'src/go.sum', 'src/**/*.go')) }}
  run: |
    printf 'value=%s\n' "$KSAIL_CACHE_KEY" >> "$GITHUB_OUTPUT"

- name: "Restore ksail binary"
  id: ksail-cache
  uses: actions/cache/restore@v4
  with:
    path: ./.cache/ksail
    key: ${{ steps.ksail-cache-key.outputs.value }}
  if: needs.build-artifact.result == 'success'

- name: "Prepare ksail binary"
  env:
    CACHE_HIT: ${{ steps.ksail-cache.outputs.cache-hit }}
  run: |
    set -Eeuo pipefail
    mkdir -p ./bin
    if [ "$CACHE_HIT" = "true" ] && [ -f ./.cache/ksail ]; then
      cp ./.cache/ksail ./bin/ksail
    else
      go build -C src -o ../bin/ksail .
      mkdir -p ./.cache
      cp ./bin/ksail ./.cache/ksail
    fi
    chmod +x ./bin/ksail
  assertions:
    - directory_exists: "./bin"
    - file_exists: "./bin/ksail"

- name: "Save ksail binary cache"
  if: steps.ksail-cache.outputs.cache-hit != 'true'
  uses: actions/cache/save@v4
  with:
    path: ./.cache/ksail
    key: ${{ steps.ksail-cache-key.outputs.value }}
```

**Failure Behavior**: If cache restore fails and rebuild fails, job must exit with non-zero status and log message.

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

## Metrics Contract (Deprecated)

> **Note:** As of Phase 6 (T028-T030), custom metrics instrumentation was removed at maintainer request. The workflow now relies on native GitHub Actions job logs for performance diagnostics instead of custom metrics collection via `$GITHUB_STEP_SUMMARY`. See research.md (lines 27-29, 106-111) for details.

## Guard Contract

```yaml
- name: "Ensure build succeeded"
  if: needs.build-artifact.result != 'success'
  run: |
    echo "Build artifact unavailable. Skipping job." >> $GITHUB_STEP_SUMMARY
    exit 1
```

**Purpose**: Prevents downstream jobs from running when the build fails; fulfills FR-006.

## Post-Run Aggregation Contract (Deprecated)

> **Note:** As of Phase 6 (T028), the `metrics-summary` aggregation job was removed entirely. Maintainers now review performance using native GitHub Actions timing data. See research.md (line 111) for details.

## Notes

- Contracts apply equally to jobs defined locally and those invoked via reusable workflows; new inputs/outputs may be required in the reusable workflow to honor them.
- Any deviation (cache misconfiguration, build failures) must fail the job and surface in the summary for fast remediation.
