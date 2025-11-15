# Implementation Plan: Flannel CNI Implementation

**Branch**: `003-flannel-cni` | **Date**: 2025-11-15 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-flannel-cni/spec.md`

## Summary

Add Flannel as a Container Network Interface (CNI) option in KSail-Go to provide reliable cluster networking compatible with standard Kubernetes setups. This follows the existing CNI architecture pattern (used by Cilium and Calico) and adds Flannel as the fourth supported CNI option using VXLAN backend exclusively for initial release. The implementation includes configuration support via `--cni Flannel` flag, installation during cluster creation, validation, graceful error handling with rollback, and comprehensive testing.

## Technical Context

**Language/Version**: Go 1.25.4+

**Primary Dependencies**:

- Kubernetes client-go (cluster interaction)
- Helm SDK (chart installation - existing pattern)
- kubectl apply (manifest installation - new for Flannel)
- testify/assert (unit testing)
- mockery v3.x (mock generation)

**Storage**: N/A (stateless CLI tool; cluster state managed by Kubernetes)

**Testing**: go test with testify assertions, mockery-generated mocks, snapshot testing for CLI output, system tests in CI

**Target Platform**: Linux, macOS, Windows (cross-platform CLI binary)

**Project Type**: Single project CLI application with package-based architecture

**Performance Goals**:

- Cluster init with Flannel: <30 seconds
- Cluster creation with Flannel: <3 minutes (distribution defaults)
- Node ready state: <60 seconds after Flannel pods running
- Pod-to-pod communication: <10ms latency

**Constraints**:

- Must follow existing CNI installer pattern (cni.InstallerBase embedding)
- Internet connectivity required for Flannel manifest download
- Kubernetes 1.20+ compatibility required
- VXLAN backend only (no custom backends)
- Distribution default cluster sizing

**Scale/Scope**:

- Add 1 new CNI option (Flannel) to existing 3 (Default, Cilium, Calico)
- Support 2 distributions: Kind and K3d
- ~500-800 lines of new code (based on Cilium installer pattern)
- 3 new files, ~10 existing files modified

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Principle-aligned gates (must all be addressed; violations documented in Complexity Tracking):

- **Simplicity (I)**: ✅ **PASS** - Reuses existing CNI installer pattern (InstallerBase embedding), no new abstractions beyond what Cilium/Calico already use. Flannel installer will be ~120 lines following existing pattern. No functions >50 lines planned.

- **Test-First (II)**: ✅ **PASS** - Tests will be written first following existing test patterns:
  - `flannel/installer_test.go` with table-driven tests (Install, Uninstall, WaitForReadiness)
  - `pkg/apis/cluster/v1alpha1/types_test.go` updates for CNIFlannel validation
  - Mock-based tests using mockery-generated mocks for Helm client and k8s interactions
  - Snapshot tests for CLI output consistency

- **Interface Discipline (III)**: ✅ **PASS** - No new interfaces required. Implements existing `installer.Installer` interface (3 methods: Install, Uninstall, SetWaitForReadinessFunc). No type switches planned.

- **Observability (IV)**: ✅ **PASS** - Follows existing CLI timing pattern:
  - `ksail cluster init --cni Flannel`: Logs CNI selection, displays timing on success
  - `ksail up`: Logs Flannel installation steps, displays component-level timing
  - Error paths: Rollback logs with diagnostic info (network failures, permission errors, version incompatibility)
  - Leverages existing `pkg/ui/timer` and `pkg/ui/notify` packages

- **Versioning (V)**: ✅ **PASS** - **MINOR** version bump (new feature, backward compatible)
  - No breaking changes to existing APIs
  - New CNI option added alongside existing ones
  - Existing configurations continue to work unchanged
  - No migration required

All gates pass. No constitutional violations.

## Project Structure

### Documentation (this feature)

```text
specs/003-flannel-cni/
├── spec.md              # Feature specification (completed)
├── plan.md              # This file (implementation plan)
├── checklists/
│   └── requirements.md  # Specification quality checklist (completed)
├── research.md          # Phase 0 output (completed)
├── data-model.md        # Phase 1 output (completed)
├── quickstart.md        # Phase 1 output (completed) - Developer quick-start guide
├── contracts/           # Phase 1 output (completed)
│   ├── flannel-installer.md  # FlannelInstaller interface contract
│   └── cni-types.md          # CNI enum extension contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - pending)
```

### Source Code (repository root)

```text
# New files to be created:
pkg/svc/installer/cni/flannel/
├── installer.go         # FlannelInstaller implementation (~120 lines)
├── installer_test.go    # Unit tests with table-driven tests (~200 lines)
└── doc.go              # Package documentation (~15 lines)

# Files to be modified:
pkg/apis/cluster/v1alpha1/
├── types.go            # Add CNIFlannel constant and update validCNIs()
└── types_test.go       # Add CNIFlannel validation tests

pkg/io/scaffolder/
└── scaffolder_test.go  # Add Flannel test cases for Kind/K3d

cmd/cluster/
└── create.go           # Add Flannel case to CNI installer factory

pkg/io/config-manager/ksail/
└── manager.go          # Update CNI handling if needed

# System test files:
.github/workflows/
└── ci.yaml             # Add Flannel to system test matrix

# Documentation:
docs/
└── cni.md              # Document Flannel CNI usage (new or update existing)

README.md               # Update CNI options list
```

**Structure Decision**: Following existing **Single Project CLI** structure with package-based architecture. All CNI installers live under `pkg/svc/installer/cni/{cni-name}/` as independent packages implementing the `installer.Installer` interface. The Flannel installer will mirror the Cilium installer structure but use kubectl apply instead of Helm for manifest installation.

## Complexity Tracking

> No constitutional violations - this section intentionally left empty.

## Planning Phase Completion

### Phase 0: Research (✅ Complete)

**Output**: [research.md](./research.md)

Resolved 6 key technical questions:

- Installation method: kubectl manifest (not Helm)
- Version strategy: /latest/ URL (rolling updates)
- Readiness detection: DaemonSet check in kube-flannel namespace
- Distribution configuration: Reuse existing disableDefaultCNI patterns
- Rollback strategy: Cluster deletion (per clarification A1)
- VXLAN configuration: Use Flannel defaults (per clarification A2)

### Phase 1: Design & Contracts (✅ Complete)

**Outputs**:

- [data-model.md](./data-model.md) - 4 entities defined with relationships and data flows
- [contracts/flannel-installer.md](./contracts/flannel-installer.md) - FlannelInstaller interface contract
- [contracts/cni-types.md](./contracts/cni-types.md) - CNI enum extension contract
- [quickstart.md](./quickstart.md) - Developer implementation guide (~3.5 hour timeline)

**Key Design Decisions**:

- FlannelInstaller embeds `cni.InstallerBase` (follows existing pattern)
- New `pkg/client/kubectl` package for manifest operations
- Install, Uninstall, SetWaitForReadinessFunc methods (standard interface)
- Table-driven tests with mockery-generated mocks
- Performance targets: <30s init, <3min cluster creation, <60s node ready

### Phase 2: Task Generation (⏳ Pending)

**Command**: `/speckit.tasks`

Will generate actionable task list from plan, contracts, and constitutional requirements.

## Next Steps

1. Run `/speckit.tasks` to generate implementation task list
2. Begin implementation following prioritized tasks (P1 → P2 → P3)
3. Follow Test-First principle: Write tests before implementation
4. Verify against success criteria from [spec.md](./spec.md)
5. Update documentation (README, docs/cni.md)
6. Submit PR with system tests

## References

- **Specification**: [spec.md](./spec.md)
- **Research**: [research.md](./research.md)
- **Data Model**: [data-model.md](./data-model.md)
- **Contracts**: [contracts/](./contracts/)
- **Quick Start**: [quickstart.md](./quickstart.md)
- **Constitution**: `.specify/memory/constitution.md`
