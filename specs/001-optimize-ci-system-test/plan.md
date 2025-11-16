# Implementation Plan: Optimize CI System-Test Build Time

**Branch**: `001-optimize-ci-system-test` | **Date**: 2025-11-16 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/001-optimize-ci-system-test/spec.md`

**Note**: This template is filled in by the `/speckit.plan` command. See `.specify/templates/commands/plan.md` for the execution workflow.

## Summary

Build and share the `ksail` binary once per workflow run and reuse warmed Go module caches in every job. A dedicated build job uploads a versioned artifact; pre-commit, reusable CI jobs, and the 11-entry system-test matrix all download the binary and rely on the unified cache. Custom metrics instrumentation has been removed per maintainer feedback, so diagnostics lean on standard job logs and artifact lineage.

## Technical Context

**Language/Version**: GitHub Actions workflow YAML orchestrating Go 1.25.4 toolchain
**Primary Dependencies**: `actions/checkout@v5`, `actions/setup-go@v6` (with cache), `actions/upload-artifact@v4`, `actions/download-artifact@v4`, `actions/cache@v4`, `pre-commit/action@v3.0.1`, `devantler-tech/reusable-workflows/.github/workflows/ci-go.yaml`
**Storage**: GitHub Actions artifact storage (5 GB per artifact, 2 GB per file) and cache backend (10 GB per repository)
**Testing**: `pre-commit` hooks, `go test ./...`, system-test matrix invoking `ksail` commands end-to-end
**Target Platform**: GitHub-hosted `ubuntu-latest` runners provisioning Kind/K3d clusters for system tests
**Project Type**: Monorepo CLI project with GitOps system tests (single backend repo)
**Performance Goals**: System-test matrix entries ≤105 seconds, total CI workflow ≤25 minutes, ≥80% Go cache hit rate, ≤10% runtime drift in unaffected jobs
**Constraints**: Shared binary must be available to reusable workflow jobs, build failure must short-circuit dependents, artifact size (~216 MB) stays within limits, parallel matrix execution preserved
**Scale/Scope**: 4 top-level jobs (pre-commit, reusable CI, system-test matrix with 11 combinations, status aggregator) plus downstream reusable workflow jobs (lint, mega-lint, test)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Principle-aligned gates (must all be addressed; violations documented in Complexity Tracking):

- **Simplicity (I)**: Introducing one `build-artifact` job and reusing existing Actions plus a single composite helper (`.github/actions/use-ksail-artifact`) keeps the workflow readable without adding unnecessary abstraction layers.
- **Test-First (II)**: Add a smoke step (`./ksail version`) in every consuming job before using the shared binary so failure cases surface immediately; write this guard before removing legacy build steps.
- **Interface Discipline (III)**: No Go interfaces added. Reusable workflow input count stays ≤5 even after adding `artifact-name`, avoiding bloated contracts and type switches.
- **Observability (IV)**: Rely on default job logs and artifact checksum tracing; guard downstream jobs with `if: needs.build-artifact.result == 'success'` to log failures and halt quickly. No custom metrics summary is maintained.
- **Versioning (V)**: Categorized as a PATCH change—CI-only optimization with no end-user or API impact.

Any gate failure must include rationale and rejected simpler alternative.

## Project Structure

### Documentation (this feature)

```text
specs/[###-feature]/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
.github/workflows/
└── ci.yaml              # Update: build job, artifact reuse, cache tuning

.github/workflows/includes/ (unchanged)
.github/workflows/templates/ (unchanged)
```

**Structure Decision**: Scope limited to `.github/workflows/ci.yaml` and supporting automation. Application source (`src/`) and other workflows remain untouched.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| *(none)* |  |  |
