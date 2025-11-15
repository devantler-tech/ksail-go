# Data Model: CNI Installer Consolidation

**Date**: 2025-11-14
**Phase**: 1 (Design)

## Overview

This refactor involves **zero data model changes**. All existing types, structs, and interfaces are preserved with identical signatures. Only package paths change.

## Existing Entities

### InstallerBase

**Location**: Currently `pkg/svc/installer/cni_helpers.go` → Moving to `pkg/svc/installer/cni/base.go`

**Purpose**: Shared base struct providing common functionality for all CNI installers.

**Fields** (unchanged):

```go
type InstallerBase struct {
    helmClient    helm.Interface
    kubeconfig    string
    context       string
    timeout       time.Duration
    waitForReady  func(context.Context) error
}
```

**Methods** (unchanged):

- `NewInstallerBase()` - Constructor
- `GetClient()` - Returns Helm client
- `GetKubeconfig()` - Returns kubeconfig path
- `GetContext()` - Returns k8s context name
- `GetTimeout()` - Returns timeout duration
- `SetWaitForReadinessFunc()` - Overrides readiness callback

**Relationships**:

- Embedded by: `CiliumInstaller`, `CalicoInstaller`
- Depends on: `helm.Interface` (from `pkg/client/helm`)
- Uses: `k8sutil.ReadinessCheck` (from `pkg/svc/installer/k8sutil`)

### HelmRepoConfig

**Location**: Currently `pkg/svc/installer/cni_helpers.go` → Moving to `pkg/svc/installer/cni/base.go`

**Purpose**: Configuration for Helm repository.

**Fields** (unchanged):

```go
type HelmRepoConfig struct {
    Name     string
    URL      string
    RepoName string
}
```

**Usage**: Passed to `InstallOrUpgradeHelmChart()` helper function.

### HelmChartConfig

**Location**: Currently `pkg/svc/installer/cni_helpers.go` → Moving to `pkg/svc/installer/cni/base.go`

**Purpose**: Configuration for Helm chart installation.

**Fields** (unchanged):

```go
type HelmChartConfig struct {
    ReleaseName     string
    ChartName       string
    Namespace       string
    RepoURL         string
    CreateNamespace bool
    SetJSONVals     map[string]string
}
```

**Usage**: Passed to `InstallOrUpgradeHelmChart()` helper function.

## CNI Installer Implementations

### CiliumInstaller

**Location**: Currently `pkg/svc/installer/cilium/installer.go` → Moving to `pkg/svc/installer/cni/cilium/installer.go`

**Purpose**: Installs Cilium CNI via Helm.

**Structure** (unchanged):

```go
type CiliumInstaller struct {
    *cni.InstallerBase
}
```

**Key Methods** (unchanged):

- `NewCiliumInstaller()` - Constructor
- `Install(ctx)` - Installs Cilium via Helm
- `Uninstall(ctx)` - Removes Cilium
- `waitForReadiness(ctx)` - Checks DaemonSet/Deployment readiness

### CalicoInstaller

**Location**: Currently `pkg/svc/installer/calico/installer.go` → Moving to `pkg/svc/installer/cni/calico/installer.go`

**Purpose**: Installs Calico CNI via Helm.

**Structure** (unchanged):

```go
type CalicoInstaller struct {
    *cni.InstallerBase
}
```

**Key Methods** (unchanged):

- `NewCalicoInstaller()` - Constructor
- `Install(ctx)` - Installs Calico via Helm
- `Uninstall(ctx)` - Removes Calico
- `waitForReadiness(ctx)` - Checks DaemonSet/Deployment readiness

## Helper Functions

### InstallOrUpgradeHelmChart

**Location**: Currently `pkg/svc/installer/cni_helpers.go` → Moving to `pkg/svc/installer/cni/base.go`

**Signature** (unchanged):

```go
func InstallOrUpgradeHelmChart(
    ctx context.Context,
    client helm.Interface,
    repoConfig HelmRepoConfig,
    chartConfig HelmChartConfig,
    timeout time.Duration,
) error
```

**Purpose**: Adds Helm repo and installs/upgrades chart.

### WaitForResourceReadiness

**Location**: Currently `pkg/svc/installer/cni_helpers.go` → Moving to `pkg/svc/installer/cni/base.go`

**Signature** (unchanged):

```go
func WaitForResourceReadiness(
    ctx context.Context,
    kubeconfig string,
    kubeContext string,
    checks []k8sutil.ReadinessCheck,
    timeout time.Duration,
    componentName string,
) error
```

**Purpose**: Waits for Kubernetes resources to become ready.

## Package Dependencies

**Current Dependencies** (preserved):

```text
pkg/svc/installer/
├── Imports:
│   ├── github.com/devantler-tech/ksail-go/pkg/client/helm
│   ├── github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil
│   ├── context
│   ├── fmt
│   └── time
```

**Post-Move Dependencies** (unchanged dependencies, only import paths updated):

```text
pkg/svc/installer/cni/
├── Imports:
│   ├── github.com/devantler-tech/ksail-go/pkg/client/helm
│   ├── github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil
│   ├── context
│   ├── fmt
│   └── time

pkg/svc/installer/cni/cilium/
├── Imports:
│   ├── github.com/devantler-tech/ksail-go/pkg/client/helm
│   ├── github.com/devantler-tech/ksail-go/pkg/svc/installer/cni  # Updated
│   ├── github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil
│   ├── context
│   ├── fmt
│   └── time

pkg/svc/installer/cni/calico/
├── Imports:
│   ├── github.com/devantler-tech/ksail-go/pkg/client/helm
│   ├── github.com/devantler-tech/ksail-go/pkg/svc/installer/cni  # Updated
│   ├── github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil
│   ├── context
│   ├── fmt
│   └── time
```

## Validation Rules

All existing validation logic preserved:

1. **Helm Client Validation**: Must be non-nil when creating installers
2. **Timeout Validation**: Must be positive duration
3. **Kubeconfig Validation**: Must be valid path
4. **Context Validation**: Must be non-empty string
5. **Readiness Check Validation**: DaemonSets/Deployments must reach "Ready" state

## State Transitions

No state management changes—installers remain stateless service objects:

1. **Initialization**: Constructor creates installer with embedded base
2. **Installation**: `Install()` adds repo → installs chart → waits for readiness
3. **Uninstallation**: `Uninstall()` removes Helm release

## Summary

This refactor is a **pure package relocation** with:

- ✅ Zero new types introduced
- ✅ Zero field additions/removals
- ✅ Zero method signature changes
- ✅ Zero validation rule changes
- ✅ Zero state management changes
- ✅ Zero dependency additions

Only package import paths change to reflect new directory structure.
