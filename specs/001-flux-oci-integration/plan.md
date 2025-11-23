# Implementation Plan: Flux OCI Integration for Automated Reconciliation

**Branch**: `001-flux-oci-integration` | **Date**: 2025-11-23 | **Spec**: `specs/001-flux-oci-integration/spec.md`
**Input**: Feature specification from `specs/001-flux-oci-integration/spec.md`

**Note**: This plan follows the KSail-Go constitution (KISS, DRY, YAGNI, interface-based design, test-first, package-first architecture, code quality gates).

## Summary

Implement Flux-based GitOps support in KSail-Go for local development clusters by:

- Installing Flux controllers during cluster bootstrap or on-demand using the Flux Go SDK.
- Provisioning a localhost-only, unauthenticated OCI registry (`registry:3`) with persistent storage and connectivity from both the cluster and the workstation.
- Building and pushing Kubernetes manifest bundles as OCI artifacts using `go-containerregistry` and a KSail-Go `workload` API.
- Generating and applying Flux `OCIRepository` and `Kustomization` resources that point to the local registry and drive automated reconciliation.
- Providing a `ksail` CLI flow to initialize, build, push, and reconcile workloads, while exposing detailed reconciliation status via Flux CRs.

## Technical Context

**Language/Version**: Go 1.25.x (per `go.mod`)
**Primary Dependencies**: Cobra (CLI), Kubernetes client-go, controller-runtime (where used), Flux Installer Go SDK, `google/go-containerregistry` for OCI artifacts, existing KSail-Go `pkg` packages
**Storage**: Local Docker volume/bind-mount for `registry:3` data (persist across cluster stop/start)
**Testing**: `go test ./...` unit/integration tests, snapshot tests under `cmd/__snapshots__`, system tests via existing CI workflows
**Target Platform**: Local Kubernetes clusters (Kind/K3d) on developer workstations (macOS/Linux), using local Docker or other container engines
**Project Type**: CLI tool with package-first architecture (`pkg/` for logic, `cmd/` for thin commands)
**Performance Goals**: End-to-end GitOps loop (artifact push → Flux reconciliation) completes within ~2 minutes; registry operations (push/pull) typically <10 seconds per operation
**Constraints**: KISS/YAGNI (local dev focus, no multi-tenant security), localhost-only registry exposure, no Helm support in this feature, re-use existing KSail-Go patterns and DI, pass `golangci-lint` and tests
**Scale/Scope**: Single-node or small local clusters, 50+ versions per artifact repository, multiple workloads per project but single local registry instance per cluster

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

- **KISS**: Plan favors a single registry, simple localhost-only networking, and Flux’s built-in CRDs. No custom controllers or complex orchestration layers introduced.
- **DRY**: Registry provisioning and Flux installation will reuse or extend existing `pkg/k8s`, `pkg/cmd`, and DI patterns instead of ad-hoc shelling out from commands.
- **YAGNI**: Feature intentionally omits HelmRelease support, advanced auth for the local registry, and multi-cluster orchestration until needed.
- **Interface-Based Design (NON-NEGOTIABLE)**: New behavior (registry management, Flux installation, workload artifact building) will be introduced behind interfaces in `pkg/` with DI wiring and mocks generated via `mockery`.
- **Test-First Development (NON-NEGOTIABLE)**: Each new package and public behavior will be accompanied by unit tests (and snapshot tests for CLI where applicable) before or alongside implementation.
- **Package-First Architecture**: Core logic will live in `pkg/` (e.g., `pkg/svc/installer/flux`, `pkg/svc/provisioner/cluster/registry`, `pkg/workload`), with `cmd/` layers delegating directly into these services.
- **Code Quality Standards (NON-NEGOTIABLE)**: All changes must pass `go test ./...` and `golangci-lint run` and integrate with existing CI (no suppression of linters without justification).

**Gate Result**: PASS – Plan adheres to all constitutional principles, with interfaces and tests as first-class deliverables.

## Project Structure

### Documentation (this feature)

```text
specs/001-flux-oci-integration/
├── spec.md          # Feature specification (input)
├── plan.md          # This file (/speckit.plan output)
├── research.md      # Phase 0 output (/speckit.plan)
├── data-model.md    # Phase 1 output (/speckit.plan)
├── quickstart.md    # Phase 1 output (/speckit.plan)
├── contracts/       # Phase 1 output (/speckit.plan)
└── tasks.md         # Phase 2 output (/speckit.tasks - not created here)
```

### Source Code (repository root)

Existing KSail-Go structure (relevant parts):

```text
cmd/
  root.go
  cluster/...
  workload/...

pkg/
  apis/
  client/
  cmd/
  di/
  io/
  k8s/
  svc/
  testutils/
  ui/
```

Planned new/updated packages for this feature (subject to refinement in Phase 1):

```text
pkg/
  svc/
    flux/           # Interfaces + implementations for Flux installation + CR management
    registry/       # Interfaces + implementations for local OCI registry lifecycle

  workload/
    oci/            # OCI artifact builder/pusher for manifest bundles (using go-containerregistry)

cmd/
  workload/
    reconcile.go    # `ksail workload reconcile` entrypoint (may already exist and be extended)
  cluster/
    flux.go         # Flux-related cluster bootstrap/on-demand install commands (or options on existing cmd)
```

**Structure Decision**: Use the existing CLI + package-first layout: all Flux, registry, and OCI logic in `pkg/svc` and `pkg/workload`, with thin `cmd/` wrappers wiring user input to services via DI. No new top-level projects or binaries.

## Complexity Tracking

No constitution violations or extra structural complexity beyond standard KSail-Go patterns are currently anticipated. This table remains empty unless later phases require justified exceptions.

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|---------------------------------------|

