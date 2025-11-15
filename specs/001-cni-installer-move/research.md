# Research: CNI Installer Consolidation

**Date**: 2025-11-14
**Phase**: 0 (Prerequisites & Discovery)

## Research Questions

### Q1: What files currently exist in pkg/svc/installer/ ?

**Finding**:

```text
pkg/svc/installer/
├── cni_helpers.go          # CNIInstallerBase + shared utilities (target: cni/base.go)
├── cni_helpers_test.go     # Tests for shared helpers (target: cni/base_test.go)
├── calico/                 # Calico CNI installer (target: cni/calico/)
│   ├── installer.go
│   └── installer_test.go
├── cilium/                 # Cilium CNI installer (target: cni/cilium/)
│   ├── installer.go
│   └── installer_test.go
└── k8sutil/               # Kubernetes utilities (unrelated to CNI, not moving)
    ├── readiness.go
    └── readiness_test.go
```

**Decision**: Move only CNI-related files. Leave `k8sutil/` package in place as it provides general Kubernetes utilities beyond CNI scope.

### Q2: Which packages import CNI installers?

**Finding**: Likely imports in:

- `cmd/cluster/` - Cluster lifecycle commands
- `pkg/provisioner/cluster/` - Cluster provisioners (Kind, K3d, EKS)
- Tests importing CNI helpers for validation

**Action Required**: Search codebase for import statements after establishing baseline.

### Q3: How are mocks generated?

**Finding**: Project uses mockery (`.mockery.yml` configuration) to generate interface mocks.

**Action Required**:

- Run `mockery` after file moves to regenerate mocks with updated import paths
- Verify mocks location (likely `pkg/svc/installer/mocks/` or adjacent)

### Q4: Are there any circular imports or hidden dependencies?

**Finding**: CNI installers depend on:

- `pkg/client/helm` - Helm client interface
- `pkg/svc/installer/k8sutil` - Readiness check utilities
- Standard library only

**Decision**: No circular imports expected. Move is safe.

## Technology Best Practices

### Go Package Organization

**Best Practice**: Group related functionality under parent packages.

**Rationale**: Go projects commonly organize by domain (e.g., `pkg/api/`, `pkg/client/`, `pkg/svc/`). CNI installers are a logical subdomain of `installer` service package.

**Applied**: Create `pkg/svc/installer/cni/` parent package with:

- Shared code at root (`cni/base.go` containing InstallerBase and utilities)
- Implementations in subdirectories (`cni/cilium/`, `cni/calico/`)

### Import Path Migration

**Best Practice**: Update all imports atomically in single commit.

**Rationale**: Partial import updates leave codebase in broken state. Atomic updates enable clean rollback via Git revert.

**Applied**: Phase 2 tasks include comprehensive import search/replace across entire codebase.

### Test Co-location

**Best Practice**: Keep test files adjacent to source files (`*_test.go` alongside `*.go`).

**Rationale**: Go convention places tests in same package or `_test` package adjacent to source.

**Applied**: Move test files with their corresponding source files during relocation.

## Integration Patterns

### Helm Client Integration

**Pattern**: Dependency injection via interface.

**Current**: CNI installers receive `helm.Interface` in constructors.

**Decision**: Preserve exact pattern—no changes to constructor signatures or Helm client usage.

### Readiness Check Integration

**Pattern**: Callback function passed to `CNIInstallerBase`.

**Current**: Each CNI implements `waitForReadiness()` method passed to base constructor.

**Decision**: Preserve callback pattern—no changes to readiness check mechanism.

### Kubernetes Client Integration

**Pattern**: Kubeconfig path + context passed as strings.

**Current**: CNI installers don't create k8s clients directly—delegate to `k8sutil` package.

**Decision**: No changes to k8s client integration—utilities remain in `k8sutil` package.

## Decisions Summary

| Decision | Rationale | Alternative Considered |
|----------|-----------|----------------------|
| Rename `cni_helpers.go` → `base.go` | Clearer semantic meaning within `cni/` package context | Keep original name (rejected: redundant "cni" prefix inside `cni/` package) |
| Keep `k8sutil/` separate | Provides general k8s utilities beyond CNI scope | Move under `cni/` (rejected: utility applies to all installers, not CNI-specific) |
| Atomic import updates | Prevents partial broken states | Gradual migration with aliases (rejected: adds complexity without benefit for internal refactor) |
| Immediate deletion of old paths | Clean break prevents confusion | Deprecation period (rejected: internal refactor doesn't need transition period per clarifications) |
| Subdirectories for implementations | Logical separation of CNI-specific code | Flat structure under `cni/` (rejected: would mix shared helpers with implementations) |

## Alternatives Considered

### Alternative 1: Flat Package Structure

**Approach**: Place all CNI code (helpers + Cilium + Calico) directly in `pkg/svc/installer/cni/` without subdirectories.

**Rejected Because**: Mixing shared helpers with CNI-specific implementations in flat structure reduces clarity. Subdirectories provide clear boundaries and make it obvious where new CNIs should be added.

### Alternative 2: Keep Existing Structure

**Approach**: Leave CNI installers at `pkg/svc/installer/cilium/` and `pkg/svc/installer/calico/`.

**Rejected Because**: Doesn't address the consolidation goal. Shared helpers remain at root `installer/` level, making it unclear they're CNI-specific. Future CNI additions lack guidance on where shared logic should live.

### Alternative 3: Create Separate `cnihelpers` Package

**Approach**: Move shared helpers to `pkg/svc/installer/cnihelpers/` instead of `pkg/svc/installer/cni/`.

**Rejected Because**: Creates unnecessary separation between helpers and implementations. Having helpers and CNIs under unified `cni/` parent provides better cohesion and discoverability.

## Blockers & Risks

### Identified Blockers

**None**: Research confirms straightforward file relocation with no architectural blockers.

### Identified Risks

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|-----------|
| Missed import references | Low | High | Comprehensive grep search for old import paths; CI will catch missed updates |
| Mock regeneration failures | Low | Medium | Test mockery locally before committing; validate mocks compile after regeneration |
| Test file relocation errors | Low | Medium | Run `go test ./...` after move; baseline comparison catches discrepancies |
| Documentation staleness | Medium | Low | Search for package path references in all `.md` files; update CONTRIBUTING.md explicitly |

## Phase 0 Completion Checklist

- ✅ Current file structure documented
- ✅ Import dependencies identified
- ✅ Mock generation process understood
- ✅ Best practices research complete
- ✅ Alternatives evaluated and documented
- ✅ No blockers identified
- ✅ Risk mitigation strategies defined

**GATE STATUS**: ✅ Proceed to Phase 1 (Design & Contracts)

