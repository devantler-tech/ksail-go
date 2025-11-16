# Data Model: Optimize CI System-Test Build Time

**Feature**: Optimize CI System-Test Build Time
**Phase**: 1 - Design
**Date**: 2025-11-16

## Overview

This feature coordinates CI workflow assets rather than runtime application data. The "data model" captures workflow artifacts, shared caches, and observability records that must stay consistent across jobs when the system-test matrix runs.

## Entities

### BinaryCacheEntry

Represents the compiled `ksail` binary stored in the GitHub Actions cache.

**Attributes**:

- `key`: string – Cache identifier (e.g., `ubuntu-ksail-bin-1.25.4-a1b2c3`)
- `paths`: array[string] – Cached filesystem locations (`.cache/ksail`)
- `sizeBytes`: integer – Size of the cached binary (~216 MB)
- `createdAt`: datetime – Timestamp when the cache entry was saved
- `originJob`: string – Job that produced the cached entry (typically `build-artifact`)

**Validation Rules**:

- `key` must encode runner OS, Go toolchain, and hashes of `src/go.mod`, `src/go.sum`, and all Go source files.
- Cache entry must exist (or be rebuilt) before downstream jobs execute smoke tests.
- `sizeBytes` must remain comfortably under the repository cache quota (10 GB aggregate).

**State Transitions**:

1. `Missing` → `Seeded` (binary compiled and saved to cache)
2. `Seeded` → `Restored` (downstream job retrieves cached binary)
3. `Restored` → `Updated` (job rebuilds binary after cache miss and re-saves entry when source hash changes)

### GoModuleCache

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
- `artifactRequired`: boolean – Whether job downloads the shared binary
- `cacheHit`: boolean – Whether Go cache restored successfully
- `status`: enum(`pending`, `running`, `success`, `failure`, `skipped`)

**Validation Rules**:

- If `artifactRequired` is true, `build-artifact` must finish with `status = success`
- `needs` list must ensure no job starts before prerequisites succeed, except `if: always()` guard for reporting job

**State Transitions**:

1. `Pending` → `Running`
2. `Running` → `Success` (all steps pass)
3. `Running` → `Failure` (smoke test or main command fails)
4. `Running` → `Skipped` (upstream failure and guard prevents execution)

## Relationships

```text
BinaryCacheEntry 1--* CIJob
  (a single cached binary is consumed by multiple jobs)

GoModuleCache 1--* CIJob
  (each job reports whether it hit the shared cache)
```

## Domain Rules

1. **Single Source Build**: Only the `build-artifact` job should seed the shared cache under normal conditions; downstream jobs may rebuild only when the cache misses and must re-save the entry afterward.
2. **Cache Consistency**: Cache keys must incorporate the Go version and `go.sum` hash to prevent stale dependency reuse.
3. **Smoke Test Guard**: Jobs consuming the artifact must execute `./ksail --version` (or similar) before cluster operations.
4. **Parallel Safety**: Cache keys include runner OS and source hashes to prevent collisions when multiple workflows execute concurrently.
5. **Failure Propagation**: If the build job fails, downstream jobs must skip execution and record the failure reason in their job logs.

## Notes

These entities map to workflow configuration and telemetry rather than runtime structs in Go. They guide how YAML changes should coordinate artifact handling, caching, and measurement across all CI jobs.
