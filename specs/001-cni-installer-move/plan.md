# Implementation Plan: CNI Installer Consolidation

**Branch**: `001-cni-installer-move` | **Date**: 2025-11-14 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/001-cni-installer-move/spec.md`

## Summary

Relocate CNI installer shared helpers (`InstallerBase`, readiness utilities, Helm configuration types) from `pkg/svc/installer/` root to `pkg/svc/installer/cni/` package, while moving Cilium and Calico installer implementations to `pkg/svc/installer/cni/cilium/` and `pkg/svc/installer/cni/calico/` subdirectories respectively. This pure structural refactor consolidates CNI-related code under a single parent package, simplifying future CNI additions while preserving all existing functionality, CLI output, and test coverage.

## Technical Context

**Language/Version**: Go 1.25.4+

**Primary Dependencies**:

- Cobra CLI (unchanged)
- Kubernetes client-go (unchanged)
- Helm Go SDK (unchanged, requires 3.8.0+)
- mockery for test mock generation

**Storage**: N/A (CLI tool with no persistent artifacts introduced)

**Testing**:

- `go test ./...` for all unit tests
- `go test ./pkg/svc/installer/cni/...` for CNI-specific tests
- mockery-generated mocks for Helm client interfaces
- CI system tests validate full cluster lifecycle (Kind/K3d/EKS)

**Target Platform**: Linux/macOS developer workstations running Kind/K3d locally or managing EKS clusters

**Project Type**: Single CLI project (cmd/, pkg/, internal/) with no new projects added

**Performance Goals**:

- CNI package unit tests: ≤ 90 seconds (QC-003)
- Full repository build: ≤ 5 minutes (existing baseline)
- CNI installation runtime: ≤ 10 minutes (unchanged from current)

**Constraints**:

- Zero functional changes—pure file relocation
- All imports updated atomically in single commit
- No new external dependencies introduced
- Existing Helm timeout and retry behavior preserved

**Scale/Scope**:

- Affects: 2 CNI installer packages (Cilium, Calico), shared base/helpers
- Touches: ~10-15 Go files (installers + tests + mocks)
- Import updates: pkg/ and cmd/ references to installer packages

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Principle I: Code Quality Discipline

**Status**: ✅ PASS

**Affected Packages**:

- `pkg/svc/installer/cni/` (new package for shared helpers)
- `pkg/svc/installer/cni/cilium/` (relocated from `pkg/svc/installer/cilium/`)
- `pkg/svc/installer/cni/calico/` (relocated from `pkg/svc/installer/calico/`)
- All packages importing CNI installers (cmd/, pkg/)

**Compliance Steps**:

- Run `gofmt`/`goimports` after file moves to ensure formatting consistency
- Execute `golangci-lint run` to verify no new warnings introduced
- Update package documentation comments to reflect new import paths
- Verify all exported APIs retain their original signatures and behavior

**Evidence Required**: Clean `golangci-lint run` output with zero errors/warnings

### Principle II: Testing Rigor

**Status**: ✅ PASS

**Test Coverage**:

- **Unit Tests**: All existing CNI installer unit tests relocated with source files
  - `pkg/svc/installer/cni/base_test.go` (relocated from `cni_helpers_test.go`)
  - `pkg/svc/installer/cni/cilium/installer_test.go` (relocated)
  - `pkg/svc/installer/cni/calico/installer_test.go` (relocated)
- **Mock Regeneration**: Run `mockery` to regenerate interface mocks with new import paths
- **Integration Tests**: CI system tests (Kind/K3d/EKS) validate full cluster lifecycle unchanged

**Failure Reproduction**:

- Import errors: `go build ./...` will fail if any references missed
- Test failures: `go test ./pkg/svc/installer/cni/...` will fail if test files not relocated correctly
- Mock issues: Tests using Helm client mocks will fail if mocks not regenerated

**Evidence Required**: `go test ./...` passes with 0 failures, all tests complete within 90 seconds for CNI packages

### Principle III: User Experience Consistency

**Status**: ✅ PASS

**CLI Surface Changes**: NONE—this is internal package restructuring only

**Affected Commands**: None directly, but CNI installers are invoked by:

- `ksail cluster up` (provisions clusters with CNI)
- `ksail cluster create` (may trigger CNI installation)

**UX Preservation**:

- All `notify` utility calls remain identical (no changes to success/error messages)
- All `timer` utility calls preserved (timing output format unchanged)
- Cobra IO streams usage unchanged (no new stdout/stderr writes)
- Help text and documentation unaffected by internal refactor

**Documentation Updates**:

- Update CONTRIBUTING.md to reference `pkg/svc/installer/cni/` as location for CNI installers
- Add inline godoc comments explaining package purpose and structure
- Update any architecture diagrams showing package layout (if present)

**Evidence Required**: `ksail cluster up` produces identical output (aside from timestamps) before/after relocation

### Principle IV: Performance & Reliability Contracts

**Status**: ✅ PASS

**Runtime Budgets**:

- CNI package unit tests: ≤ 90 seconds (measured via `go test -v ./pkg/svc/installer/cni/...`)
- Full repository build: ≤ 5 minutes (existing CI baseline)
- CNI installation: ≤ 10 minutes with progress feedback (unchanged behavior)

**Progress Feedback**: Preserved via existing timer utilities—no changes to user-visible progress indicators

**Failure Handling**:

- Existing error handling unchanged (preserve all `fmt.Errorf` wrapping chains)
- Git revert recovery strategy if relocation breaks builds (no automated rollback needed)
- Helm timeout/retry behavior preserved exactly as-is

**Reliability Contracts**:

- Zero functional changes—pure structural refactor
- All existing Helm chart verification and RBAC requirements preserved (QC-004)
- No changes to concurrency, backoff, or retry logic

**Evidence Required**:

- CI build/test timings remain within established baselines
- `go test` timing output showing CNI tests complete in ≤ 90s
- Successful cluster creation demonstrating unchanged CNI installation behavior

**GATE STATUS**: ✅ All four principles satisfied—proceed to Phase 0 research

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

**Current Structure** (before relocation):

```text
pkg/svc/installer/
├── cni_helpers.go          # InstallerBase + shared utilities
├── cni_helpers_test.go     # Tests for shared helpers
├── calico/                 # Calico CNI installer
│   ├── installer.go
│   └── installer_test.go
└── cilium/                 # Cilium CNI installer
    ├── installer.go
    └── installer_test.go
```

**Target Structure** (after relocation):

```text
pkg/svc/installer/cni/
├── base.go                 # InstallerBase and shared utilities (renamed from cni_helpers.go)
├── base_test.go            # Tests (renamed from cni_helpers_test.go)
├── doc.go                  # Package documentation
├── calico/                 # Calico CNI installer
│   ├── installer.go
│   └── installer_test.go
└── cilium/                 # Cilium CNI installer
    ├── installer.go
    └── installer_test.go
```

**Structure Decision**: Single project (Go CLI) with package consolidation. All CNI-related code grouped under `pkg/svc/installer/cni/` parent package for logical cohesion. Shared helpers live at package root (`cni/base.go`), while CNI-specific implementations remain in subdirectories (`cni/cilium/`, `cni/calico/`).

**Files to Move**:

1. `pkg/svc/installer/cni_helpers.go` → `pkg/svc/installer/cni/base.go`
2. `pkg/svc/installer/cni_helpers_test.go` → `pkg/svc/installer/cni/base_test.go`
3. `pkg/svc/installer/cilium/` → `pkg/svc/installer/cni/cilium/`
4. `pkg/svc/installer/calico/` → `pkg/svc/installer/cni/calico/`

**Import Path Changes**:

- Old: `github.com/devantler-tech/ksail-go/pkg/svc/installer`
- New (shared): `github.com/devantler-tech/ksail-go/pkg/svc/installer/cni`
- New (Cilium): `github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/cilium`
- New (Calico): `github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/calico`

## Complexity Tracking

**Status**: No violations—all Constitution principles satisfied.

This is a pure structural refactor with:

- Zero new dependencies
- Zero functional changes
- Zero new abstractions or patterns introduced
- All existing tests preserved and relocated

No complexity justification required.

## Phase 0: Research & Prerequisites

**Objective**: Validate assumptions about current package structure and dependencies.

### Research Tasks

1. **Inventory Current Files**
   - List all files in `pkg/svc/installer/` to confirm structure matches plan
   - Identify any additional files beyond installers/helpers (e.g., mocks, testdata)
   - Document any unexpected dependencies or circular imports

2. **Scan Import References**
   - Search codebase for all imports of `pkg/svc/installer` package
   - Identify files importing Cilium/Calico installers directly
   - Document cmd/ and pkg/ packages affected by import changes

3. **Verify Mock Generation**
   - Confirm mockery configuration (`.mockery.yml`) for installer interfaces
   - Identify which mocks need regeneration after move
   - Test mock regeneration locally before committing changes

4. **Review Test Dependencies**
   - Check for test files importing CNI helpers or installers
   - Verify no hardcoded package paths in test fixtures or data
   - Confirm CI test commands target correct paths after move

### Expected Findings

- **File Count**: ~10-15 Go files (installers, tests, helpers, mocks)
- **Import References**: Primarily in `cmd/cluster/` commands and `pkg/provisioner/` packages
- **Mock Files**: Helm client mocks likely in `pkg/svc/installer/mocks/` or adjacent
- **Test Coverage**: Existing unit tests for base helpers and individual CNI installers

### Outputs

- `research.md` documenting findings and confirming no blockers
- List of all files requiring import path updates
- Mockery regeneration commands validated locally

**GATE**: Research complete—proceed to Phase 1 if no blockers discovered

## Phase 1: Design & Contracts

**Objective**: Define concrete file operations and validation strategy.

### Data Model

**Entities** (unchanged—no data model changes in this refactor):

- `InstallerBase`: Shared base struct for CNI installers
- `HelmRepoConfig`: Helm repository configuration
- `HelmChartConfig`: Helm chart installation parameters

**Package Documentation** (new):

Create `pkg/svc/installer/cni/doc.go`:

```go
// Package cni provides Container Network Interface (CNI) installer implementations
// and shared utilities for Kubernetes cluster networking setup.
//
// Structure:
//   - base.go: InstallerBase and utility functions for Helm chart installation and readiness checks
//   - calico/: Calico CNI installer implementation
//   - cilium/: Cilium CNI installer implementation
//
// Adding a new CNI:
// 1. Create a subdirectory under pkg/svc/installer/cni/
// 2. Implement the installer.Installer interface
// 3. Embed InstallerBase for shared Helm client and readiness logic
// 4. Add unit tests following existing patterns
package cni
```

### API Contracts

**No API changes**—this is internal restructuring only. All exported functions and types retain identical signatures:

**Preserved Contracts**:

- `InstallerBase` methods unchanged
- `InstallOrUpgradeHelmChart()` signature identical
- `WaitForResourceReadiness()` signature identical
- Cilium/Calico installer constructors unchanged
- All installer interface implementations preserved

### Quickstart

**For Contributors Adding New CNIs**:

```bash
# 1. Create new CNI package
mkdir -p pkg/svc/installer/cni/mycni

# 2. Create installer implementation
cat > pkg/svc/installer/cni/mycni/installer.go << 'EOF'
package mycniinstaller

import (
    "context"
    "time"
    "github.com/devantler-tech/ksail-go/pkg/client/helm"
    "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"
)

type MyCNIInstaller struct {
    *cni.InstallerBase
}

func NewMyCNIInstaller(
    client helm.Interface,
    kubeconfig, context string,
    timeout time.Duration,
) *MyCNIInstaller {
    installer := &MyCNIInstaller{}
    installer.InstallerBase = cni.NewInstallerBase(
        client,
        kubeconfig,
        context,
        timeout,
        installer.waitForReadiness,
    )
    return installer
}

func (m *MyCNIInstaller) Install(ctx context.Context) error {
    // Implementation here
    return nil
}

func (m *MyCNIInstaller) waitForReadiness(ctx context.Context) error {
    // Readiness checks here
    return nil
}
EOF

# 3. Add unit tests
cat > pkg/svc/installer/cni/mycni/installer_test.go << 'EOF'
package mycniinstaller

import (
    "testing"
)

func TestMyCNIInstaller_Install(t *testing.T) {
    // Test implementation
}
EOF

# 4. Run tests
go test ./pkg/svc/installer/cni/mycni/...

# 5. Generate mocks if needed
mockery

# 6. Verify build
go build ./...
```

### Validation Strategy

**Pre-Move Validation**:

1. Capture baseline test results: `go test ./pkg/svc/installer/... > pre-move-tests.txt`
2. Capture baseline lint output: `golangci-lint run > pre-move-lint.txt`
3. Document current import paths: `grep -r "pkg/svc/installer" . --include="*.go" > pre-move-imports.txt`

**Post-Move Validation**:

1. **Build Check**: `go build ./...` must succeed with zero errors
2. **Test Check**: `go test ./pkg/svc/installer/cni/...` must pass all tests
3. **Lint Check**: `golangci-lint run` must show zero new warnings
4. **Import Check**: `grep -r "pkg/svc/installer/cilium\|pkg/svc/installer/calico" . --include="*.go"` must return zero results (old paths removed)
5. **Mock Check**: Verify mocks regenerated with `git status | grep "mocks/"`
6. **Timing Check**: `go test -v ./pkg/svc/installer/cni/...` completes in ≤ 90 seconds

**Integration Validation**:

1. **Local Cluster Test**: `ksail cluster up` with Kind + Cilium (default CNI)
2. **Output Verification**: Compare CLI output with baseline (timestamps excluded)
3. **CI Validation**: Push to feature branch and verify CI system tests pass

### Phase 1 Deliverables

- `data-model.md` (minimal—no data model changes)
- `quickstart.md` with contributor guidance
- Validation checklist for post-move verification

**GATE**: Constitution Re-check—confirm all principles still satisfied after design phase

### Constitution Re-check (Post-Design)

- ✅ **Principle I**: File operations defined, lint/format steps included
- ✅ **Principle II**: Test relocation strategy documented, validation commands specified
- ✅ **Principle III**: No CLI changes, documentation updates identified
- ✅ **Principle IV**: Timing budgets confirmed, no performance impact expected

**GATE STATUS**: ✅ Proceed to Phase 2 (task breakdown via `/speckit.tasks`)

## Phase 2: Task Breakdown

**Note**: Phase 2 task breakdown is generated by `/speckit.tasks` command, not `/speckit.plan`.

Task file will be created at: `specs/001-cni-installer-move/tasks.md`

**Expected Task Categories**:

1. **Preparation Tasks**: Baseline capture, research validation
2. **File Move Tasks**: Git mv operations for each file/directory
3. **Import Update Tasks**: Update import statements across codebase
4. **Mock Regeneration Tasks**: Run mockery with new paths
5. **Test Tasks**: Validate unit/integration/system tests
6. **Documentation Tasks**: Update CONTRIBUTING.md, add package docs
7. **Cleanup Tasks**: Remove old directories, verify no stale references

## Implementation Summary

**Implementation Approach**: Pure structural refactor—atomic file moves + import updates in single commit.

**Key Technical Decisions**:

1. Shared helpers → `pkg/svc/installer/cni/base.go` (renamed for clarity)
2. CNI implementations → subdirectories (`cni/cilium/`, `cni/calico/`)
3. Import updates applied atomically (no transition period)
4. Mocks regenerated after move to reflect new paths
5. Zero functional changes—all existing behavior preserved

**Risk Mitigation**:

- Pre-move baseline capture enables easy rollback via Git revert
- Comprehensive validation suite catches missed imports/broken tests
- CI system tests provide integration-level regression detection
- Constitution compliance ensures quality standards maintained

**Success Metrics** (from spec.md):

- SC-001: CNI tests pass in ≤ 90 seconds ✅
- SC-002: Build pipeline succeeds with zero import errors ✅
- SC-003: Documentation references new canonical location ✅
- SC-004: CI timings remain within baselines ✅

**Next Steps**:

1. Review this plan with stakeholders
2. Run `/speckit.tasks` to generate detailed task breakdown
3. Execute tasks following constitution principles
4. Submit PR with evidence of all success criteria met
