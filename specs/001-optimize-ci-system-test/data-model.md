# Data Model: Optimize CI System-Test Build Time

**Feature**: Optimize CI System-Test Build Time
**Phase**: 1 - Design
**Date**: 2025-11-16

## Overview

This feature coordinates CI workflow assets rather than runtime application data. The "data model" captures workflow artifacts, shared caches, and observability records that must stay consistent across jobs when the system-test matrix runs.

## Entities

### BuildArtifact

Represents the compiled `ksail` binary produced once per workflow run.

**Attributes**:

- `name`: string – Artifact identifier (e.g., `ksail-${{ github.run_id }}`)
- `version`: string – Source revision (commit SHA) embedded in the artifact metadata
- `sizeBytes`: integer – Artifact file size (expected ~216 MB)
- `checksum`: string – SHA256 digest stored alongside the artifact
- `createdAt`: datetime – Timestamp when the build job finished
- `retentionDays`: integer – Artifact retention (default 90 days)

**Validation Rules**:

- `sizeBytes` must be < 2 GB (GitHub Actions per-file limit)
- `checksum` must match the downloaded file before consumers execute it
- Artifact must exist before any downstream job runs smoke tests

**State Transitions**:

1. `Pending` → `Built` (binary compiled)
2. `Built` → `Uploaded` (artifact uploaded successfully)
3. `Uploaded` → `Consumed` (downstream job verifies smoke test)
4. `Consumed` → `Expired` (auto-deletion after retention window)

### GoCache

Represents the shared Go module and build cache.

**Attributes**:

- `key`: string – Cache key (e.g., `ubuntu-go-${hash(src/go.sum)}`)
- `paths`: array[string] – Cached directories (`~/.cache/go-build`, `~/go/pkg/mod`)
- `hit`: boolean – Whether cache was restored successfully
- `toolchain`: string – Go toolchain version (e.g., `1.25.4`)
- `lastUpdated`: datetime – Last time cache was saved

**Validation Rules**:

- `key` must include OS and go.sum hash to avoid collisions between branches
- `toolchain` must match the version used in the build job
- If `hit` is false, the job must log a miss and continue without failure

**State Transitions**:

1. `Warm` (cache restored) → `Updated` (dependencies downloaded, cache saved)
2. `Cold` (miss) → `Updated`
3. `Updated` → `Invalid` (go.sum or toolchain change)

### CIJob

Represents a GitHub Actions job participating in the workflow.

**Attributes**:

- `id`: string – Job identifier (`pre-commit`, `build-artifact`, `system-test`)
- `needs`: array[string] – Upstream job dependencies
- `durationSeconds`: integer – Total run time captured via metrics step
- `artifactRequired`: boolean – Whether job downloads the shared binary
- `cacheHit`: boolean – Whether Go cache restored successfully
- `status`: enum(`pending`, `running`, `success`, `failure`, `skipped`)
- `summaryPath`: string – Path to job summary file appended during run

**Validation Rules**:

- If `artifactRequired` is true, `build-artifact` must finish with `status = success`
- Jobs must log duration and cache status before completion
- `needs` list must ensure no job starts before prerequisites succeed, except `if: always()` guard for reporting job

**State Transitions**:

1. `Pending` → `Running`
2. `Running` → `Success` (all steps pass)
3. `Running` → `Failure` (smoke test or main command fails)
4. `Running` → `Skipped` (upstream failure and guard prevents execution)

### MetricsRecord

Represents the aggregated telemetry appended to `$GITHUB_STEP_SUMMARY`.

**Attributes**:

- `jobId`: string – ID of the emitting job
- `artifactVersion`: string – Artifact checksum or commit SHA
- `durationSeconds`: integer – Derived `(end - start)`
- `cacheStatus`: enum(`hit`, `miss`, `n/a`)
- `notes`: string – Additional diagnostics (e.g., "go mod download skipped")

**Validation Rules**:

- `jobId` must align with an existing `CIJob`
- `durationSeconds` must be > 0
- `cacheStatus` must reflect the actual cache result (from `restore` or `save` step outputs)

**State Transitions**:

1. `Collected` (data available in job) → `Published` (written to summary)
2. `Published` → `Reviewed` (maintainer inspects summary)

## Relationships

```text
BuildArtifact 1--* CIJob
  (a single artifact is consumed by multiple jobs)

GoCache 1--* CIJob
  (each job reports whether it hit the shared cache)

CIJob 1--1 MetricsRecord
  (each job must emit exactly one metrics summary)
```

## Domain Rules

1. **Single Source Build**: Only the `build-artifact` job may create the shared binary; all other jobs must treat the artifact as read-only consumption.
2. **Cache Consistency**: Cache keys must incorporate the Go version and `go.sum` hash to prevent stale dependency reuse.
3. **Smoke Test Guard**: Jobs consuming the artifact must execute `./ksail version` (or similar) before cluster operations.
4. **Metrics Requirement**: Every job must emit a `MetricsRecord` entry; missing summaries trigger follow-up investigation.
5. **Parallel Safety**: Artifact names include `github.run_id` to prevent cross-run contamination when multiple workflows execute concurrently.
6. **Failure Propagation**: If the build job fails, downstream jobs must skip execution and record the failure reason in their metrics output.

## Notes

These entities map to workflow configuration and telemetry rather than runtime structs in Go. They guide how YAML changes should coordinate artifact handling, caching, and measurement across all CI jobs.
