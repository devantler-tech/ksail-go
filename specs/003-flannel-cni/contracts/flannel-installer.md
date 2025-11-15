# Contract: Flannel Installer Interface

**Package**: `pkg/svc/installer/cni/flannel`
**Interface**: `installer.Installer` (existing)
**Date**: 2025-11-15

## Interface Implementation

The `FlannelInstaller` implements the existing `installer.Installer` interface:

```go
// Installer defines the contract for component installers
type Installer interface {
    Install(ctx context.Context) error
    Uninstall(ctx context.Context) error
    SetWaitForReadinessFunc(waitFunc func(context.Context) error)
}
```

## Method Contracts

### Install

```go
func (f *FlannelInstaller) Install(ctx context.Context) error
```

**Purpose**: Installs Flannel CNI by applying the official manifest and waiting for readiness

**Preconditions**:

- Context must not be cancelled
- Kubernetes cluster must be accessible
- Internet connectivity available to download manifest
- Default CNI must be disabled in cluster configuration
- Installer must have valid kubeconfig and context

**Postconditions (Success)**:

- Flannel manifest applied to cluster
- `kube-flannel` namespace created
- `kube-flannel-ds` DaemonSet running on all nodes
- All Flannel pods in Ready state
- Nodes have network connectivity configured
- Function returns `nil`

**Postconditions (Failure)**:

- Returns wrapped error with context
- Error message includes diagnostic information
- Cluster may be partially configured (caller handles rollback per FR-011a)
- No state persisted in installer (stateless)

**Error Conditions**:

- `ErrNetworkUnavailable`: Cannot reach manifest URL or cluster
- `ErrPermissionDenied`: Insufficient Kubernetes RBAC permissions
- `ErrTimeout`: Readiness check exceeded timeout duration
- `ErrInvalidManifest`: Manifest YAML is malformed or invalid
- `ErrResourceConflict`: Namespace or resource already exists in unexpected state

**Performance**:

- Expected duration: 30-90 seconds (varies by network and cluster size)
- Timeout: Configurable via cluster configuration (default: 5 minutes)
- Non-blocking: Uses context for cancellation

**Example Usage**:

```go
installer := flannel.NewFlannelInstaller(
    kubectlClient,
    "/home/user/.kube/config",
    "kind-local",
    5*time.Minute,
)

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := installer.Install(ctx)
if err != nil {
    // Handle error - caller may trigger rollback
    return fmt.Errorf("flannel installation failed: %w", err)
}
```

### Uninstall

```go
func (f *FlannelInstaller) Uninstall(ctx context.Context) error
```

**Purpose**: Removes Flannel CNI components from the cluster

**Preconditions**:

- Context must not be cancelled
- Kubernetes cluster must be accessible
- Flannel resources may or may not exist (idempotent)

**Postconditions (Success)**:

- Flannel DaemonSet deleted
- `kube-flannel` namespace resources removed
- Namespace may remain (deletion is best-effort)
- Function returns `nil`

**Postconditions (Failure)**:

- Returns wrapped error with context
- Partial cleanup may have occurred
- Cluster state may be inconsistent (caller should handle)

**Error Conditions**:

- `ErrNetworkUnavailable`: Cannot reach cluster
- `ErrPermissionDenied`: Insufficient RBAC permissions to delete resources
- `ErrTimeout`: Deletion exceeded timeout duration

**Performance**:

- Expected duration: 10-30 seconds
- Timeout: Same as Install timeout
- Non-blocking: Uses context for cancellation

**Example Usage**:

```go
ctx := context.Background()

err := installer.Uninstall(ctx)
if err != nil {
    log.Warnf("flannel uninstall failed: %v", err)
    // Continue with cluster deletion even if uninstall fails
}
```

### SetWaitForReadinessFunc

```go
func (f *FlannelInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error)
```

**Purpose**: Override the default readiness check function (primarily for testing)

**Preconditions**:

- Installer must be instantiated
- Called before `Install()` to take effect

**Postconditions**:

- Readiness function replaced with provided function
- If `waitFunc` is `nil`, restores default readiness check
- No validation of function behavior

**Error Conditions**:

- None (method does not return error)

**Example Usage (Testing)**:

```go
// Mock readiness check for testing
mockWait := func(ctx context.Context) error {
    return nil // Simulate immediate readiness
}

installer.SetWaitForReadinessFunc(mockWait)

// Test Install without actual readiness wait
err := installer.Install(ctx)
assert.NoError(t, err)
```

## Internal Method Contracts

### waitForReadiness (private)

```go
func (f *FlannelInstaller) waitForReadiness(ctx context.Context) error
```

**Purpose**: Wait for Flannel DaemonSet to become ready on all nodes

**Preconditions**:

- Flannel manifest has been applied
- Context is valid

**Postconditions (Success)**:

- All Flannel pods are running and ready
- DaemonSet reports desired == ready == available
- Returns `nil`

**Postconditions (Failure)**:

- Timeout reached before readiness
- Returns error with diagnostic information

**Readiness Criteria**:

```go
checks := []k8s.ReadinessCheck{
    {
        Type:      "daemonset",
        Namespace: "kube-flannel",
        Name:      "kube-flannel-ds",
    },
}
```

- **DaemonSet Status**:
  - `status.desiredNumberScheduled == status.numberReady`
  - `status.numberReady == status.numberAvailable`
  - All pods have passed readiness probes

## Constructor Contract

### NewFlannelInstaller

```go
func NewFlannelInstaller(
    client kubectl.Interface,
    kubeconfig, context string,
    timeout time.Duration,
) *FlannelInstaller
```

**Purpose**: Factory function to create a new Flannel installer instance

**Parameters**:

- `client`: kubectl client interface for applying manifests (must not be nil)
- `kubeconfig`: Path to kubeconfig file (must be valid file path)
- `context`: Kubernetes context name (must exist in kubeconfig)
- `timeout`: Maximum wait duration for operations (must be positive)

**Returns**:

- Fully initialized `*FlannelInstaller` with embedded `InstallerBase`
- Default readiness function configured
- Ready to call `Install()`

**Panics**: If client is nil (defensive programming)

**Example**:

```go
kubectlClient, err := kubectl.New(kubeconfig, context)
if err != nil {
    return nil, fmt.Errorf("create kubectl client: %w", err)
}

installer := flannel.NewFlannelInstaller(
    kubectlClient,
    kubeconfig,
    context,
    5*time.Minute,
)
```

## Type Constraints

### kubectl.Interface

```go
type Interface interface {
    Apply(ctx context.Context, manifestURL string) error
}
```

**Contract**:

- `Apply` must fetch manifest from URL and apply to cluster
- Must handle HTTP/HTTPS URLs
- Must validate YAML before applying
- Must return descriptive errors on failure
- Must be mockable for testing

## Testing Contract

### Unit Tests Required

```go
func TestFlannelInstaller_Install(t *testing.T)
func TestFlannelInstaller_Install_NetworkError(t *testing.T)
func TestFlannelInstaller_Install_Timeout(t *testing.T)
func TestFlannelInstaller_Uninstall(t *testing.T)
func TestFlannelInstaller_SetWaitForReadinessFunc(t *testing.T)
func TestNewFlannelInstaller(t *testing.T)
```

### Table-Driven Test Structure

```go
tests := []struct {
    name          string
    mockSetup     func(*MockKubectlClient)
    wantErr       bool
    errorContains string
}{
    {
        name: "successful installation",
        mockSetup: func(m *MockKubectlClient) {
            m.On("Apply", mock.Anything, flannelManifestURL).Return(nil)
        },
        wantErr: false,
    },
    {
        name: "network error during apply",
        mockSetup: func(m *MockKubectlClient) {
            m.On("Apply", mock.Anything, flannelManifestURL).
                Return(fmt.Errorf("network unavailable"))
        },
        wantErr:       true,
        errorContains: "network unavailable",
    },
    // ... more cases
}
```

### Mock Requirements

- `MockKubectlClient` generated via mockery
- `MockInstallerBase` for testing readiness overrides
- Test fixtures for manifest YAML

## Backward Compatibility

- **No breaking changes**: Implements existing `installer.Installer` interface
- **Additive only**: New installer type, existing installers unchanged
- **Optional feature**: Flannel use is opt-in via configuration
- **Version**: MINOR bump (new feature, backward compatible)

## Dependencies

### Required Packages

- `context` (standard library)
- `fmt` (standard library)
- `time` (standard library)
- `github.com/devantler-tech/ksail-go/pkg/client/kubectl` (NEW)
- `github.com/devantler-tech/ksail-go/pkg/k8s`
- `github.com/devantler-tech/ksail-go/pkg/svc/installer`
- `github.com/devantler-tech/ksail-go/pkg/svc/installer/cni`

### External Dependencies

- Flannel manifest URL: `https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml`
- Kubernetes cluster with API access
- Internet connectivity for manifest download
