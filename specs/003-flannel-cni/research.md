# Research: Flannel CNI Implementation

**Phase**: 0 - Outline & Research
**Date**: 2025-11-15
**Status**: Complete

## Research Questions Resolved

### 1. Flannel Installation Method

**Question**: How should Flannel be installed - via Helm chart or kubectl manifest?

**Decision**: kubectl manifest (kube-flannel.yml from official Flannel GitHub releases)

**Rationale**:

- Flannel's official installation method is via kubectl manifest application
- Helm chart exists but is community-maintained and less commonly used
- Official documentation recommends manifest installation
- Simpler approach aligns with KISS principle
- Manifest URL pattern: `https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml`

**Alternatives Considered**:

- Helm chart: Rejected because not officially maintained, adds unnecessary complexity
- Manual CNI plugin installation: Rejected because manifest handles everything

**Implementation Approach**: Use kubectl apply with manifest URL (similar to how other projects install Flannel)

### 2. Flannel Manifest Source and Versioning

**Question**: Which Flannel version should be used and how to ensure compatibility?

**Decision**: Use `/latest/download/kube-flannel.yml` for automatic latest stable version

**Rationale**:

- Flannel maintains good backward compatibility
- `/latest` endpoint always points to current stable release
- Kubernetes 1.20+ compatibility requirement covered by modern Flannel releases
- Reduces maintenance burden (no hardcoded versions to update)

**Alternatives Considered**:

- Pinned version (e.g., v0.25.0): Rejected because requires manual updates, version management overhead
- Version detection from cluster: Rejected as over-engineering for initial release

**Implementation Approach**: Hard-code manifest URL with `/latest/` path

### 3. Readiness Check Strategy

**Question**: How to verify Flannel installation success?

**Decision**: Check for DaemonSet `kube-flannel-ds` in `kube-flannel` namespace to be ready

**Rationale**:

- Flannel runs as a DaemonSet (one pod per node)
- DaemonSet readiness indicates network plugin is active
- Namespace is `kube-flannel` (not `kube-system` like other CNIs)
- Follows existing pattern from Cilium installer (DaemonSet + Deployment checks)

**Alternatives Considered**:

- Pod-level checks: Rejected because DaemonSet status is sufficient
- Network connectivity tests: Rejected as too complex for installer-level validation (belongs in E2E tests)

**Implementation Approach**: Use existing `installer.WaitForResourceReadiness()` with DaemonSet check for `kube-flannel-ds` in `kube-flannel` namespace

### 4. Distribution-Specific Configuration

**Question**: How to handle Kind vs K3d differences for Flannel?

**Decision**: Both require disableDefaultCNI, no other distribution-specific changes needed

**Rationale**:

- Kind: Needs `disableDefaultCNI: true` in kind-config.yaml (same as Cilium)
- K3d: Needs `--flannel-backend=none` flag to disable built-in Flannel (existing behavior)
- No additional Flannel-specific configuration required
- Scaffolder already handles CNI disabling logic

**Alternatives Considered**:

- Distribution-specific Flannel manifests: Rejected because unnecessary, single manifest works universally
- Custom network CIDR configuration: Rejected as out of scope (standard Flannel config sufficient)

**Implementation Approach**: Reuse existing CNI disabling logic in scaffolder, no new distribution handling needed

### 5. Error Handling and Rollback

**Question**: How to implement graceful failure and rollback as specified?

**Decision**: Catch installation errors, attempt cluster deletion, display diagnostic message

**Rationale**:

- Aligns with clarification requirement (fail gracefully with rollback)
- Prevents partial/broken cluster states
- Existing provisioner patterns support cluster deletion on failure
- Error messages can include specific failure reasons (network, permissions, version)

**Alternatives Considered**:

- Retry logic: Rejected per clarification (not requested, adds complexity)
- Partial rollback: Rejected because full cluster cleanup is simpler and safer

**Implementation Approach**: Wrap Flannel installation in error handler that triggers cluster deletion and formats diagnostic error messages

### 6. VXLAN Backend Configuration

**Question**: Does Flannel need explicit VXLAN backend configuration?

**Decision**: No explicit configuration needed - VXLAN is Flannel's default backend

**Rationale**:

- Flannel defaults to VXLAN backend without any configuration
- Official kube-flannel.yml manifest uses VXLAN by default
- No need to modify manifest or add custom ConfigMap
- Simplest approach per KISS principle

**Alternatives Considered**:

- Explicit VXLAN configuration in manifest: Rejected as unnecessary (default behavior)
- Backend selection flag: Rejected as out of scope per clarification

**Implementation Approach**: Use unmodified official manifest; VXLAN is automatic

## Technology Stack

### Core Technologies

- **Kubernetes client-go**: Existing dependency for cluster interaction
- **kubectl apply**: New usage pattern - apply manifest from URL
- **Flannel**: CNI plugin, version-agnostic (latest stable via GitHub releases)

### Installation Pattern

```go
// Pseudocode pattern
type FlannelInstaller struct {
    *cni.InstallerBase
}

func (f *FlannelInstaller) Install(ctx context.Context) error {
    // 1. Apply Flannel manifest via kubectl
    manifestURL := "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml"
    err := f.applyManifest(ctx, manifestURL)
    if err != nil {
        return rollbackCluster(err) // FR-011a
    }

    // 2. Wait for readiness
    return f.WaitForReadiness(ctx)
}

func (f *FlannelInstaller) waitForReadiness(ctx context.Context) error {
    checks := []k8s.ReadinessCheck{
        {Type: "daemonset", Namespace: "kube-flannel", Name: "kube-flannel-ds"},
    }
    return installer.WaitForResourceReadiness(ctx, kubeconfig, context, checks, timeout, "flannel")
}
```

### Key Differences from Cilium Installer

| Aspect | Cilium | Flannel |
|--------|--------|---------|
| Installation Method | Helm chart | kubectl manifest |
| Namespace | kube-system | kube-flannel |
| Primary Resource | DaemonSet + Deployment | DaemonSet only |
| Configuration | Helm values | None (default VXLAN) |
| Dependencies | Helm SDK | kubectl/client-go |

## Implementation Dependencies

### External Dependencies

1. **Flannel GitHub Releases**: Manifest source (<https://github.com/flannel-io/flannel>)
2. **Internet Connectivity**: Required to download manifest (documented assumption)
3. **Kubernetes 1.20+**: Minimum cluster version (documented dependency)

### Internal Dependencies

1. **pkg/client/kubectl** (NEW): Need kubectl manifest apply capability
   - May need to add kubectl client abstraction if not exists
   - Alternative: Use client-go dynamic client to apply manifest
2. **pkg/svc/installer/cni.InstallerBase**: Existing base class for CNI installers
3. **pkg/svc/installer.WaitForResourceReadiness**: Existing readiness checker
4. **pkg/k8s.ReadinessCheck**: Existing struct for resource checks

### New Package Required

**Decision**: Create `pkg/client/kubectl` package for manifest application

**Rationale**:

- Flannel uses manifest, not Helm (unlike Cilium/Calico)
- Reusable for future manifest-based installations
- Clean separation of concerns (kubectl operations isolated)
- Interface-based design for testability

**Interface Design**:

```go
package kubectl

// Interface defines kubectl operations
type Interface interface {
    Apply(ctx context.Context, manifestURL string) error
}
```

## Best Practices Identified

### From Kubernetes CNI Documentation

1. **Disable Default CNI**: Always disable distribution default CNI when using custom CNI
2. **Namespace Isolation**: CNI components should run in dedicated namespace
3. **DaemonSet Pattern**: CNI must run on every node (DaemonSet is standard)
4. **Readiness Checks**: Verify CNI is running before declaring cluster ready

### From Flannel Project

1. **VXLAN Default**: Use VXLAN for universal compatibility (official recommendation)
2. **No Manual Configuration**: Default settings work for vast majority of use cases
3. **Latest Stable**: Use /latest/ endpoint for automatic updates
4. **Standard Ports**: Flannel uses UDP 8285 (VXLAN) - no configuration needed

### From Existing KSail-Go Patterns

1. **Interface Embedding**: Embed `*cni.InstallerBase` for shared functionality
2. **Table-Driven Tests**: Use table-driven test pattern for all scenarios
3. **Mock-Based Testing**: Generate mocks with mockery for dependencies
4. **Error Wrapping**: Wrap errors with context using `fmt.Errorf("context: %w", err)`

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Flannel manifest URL changes | High | Use /latest/ endpoint; monitor Flannel releases; fallback to version pins if needed |
| Network unavailable during install | Medium | Clear error message per FR-011, rollback per FR-011a, document internet requirement |
| Namespace conflicts | Low | Standard `kube-flannel` namespace unlikely to conflict; check exists in readiness |
| Incompatible K8s version | Low | Document K8s 1.20+ requirement; Flannel broadly compatible |
| kubectl client not available | Medium | Create pkg/client/kubectl abstraction using client-go dynamic client |

## Open Questions

**None** - All research questions resolved per specification requirements and clarifications.
