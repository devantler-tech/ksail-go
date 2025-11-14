# Quickstart: Adding New CNI Installers

**Last Updated**: 2025-11-14
**Applies to**: KSail Go v0.x after CNI consolidation

## Overview

This guide shows how to add a new Container Network Interface (CNI) installer to KSail Go. All CNI-related code lives under `pkg/svc/installer/cni/` to keep implementations organized and maintainable.

## Prerequisites

- Go 1.23.9+ installed
- Familiarity with Helm chart installation
- Understanding of Kubernetes readiness checks
- Access to CNI's Helm chart repository

## Project Structure

```text
pkg/svc/installer/cni/
├── base.go                 # CNIInstallerBase and shared utilities (embed this in your installer)
├── base_test.go            # Tests for base functionality
├── doc.go                  # Package documentation
├── calico/                 # Example: Calico implementation
│   ├── installer.go
│   └── installer_test.go
├── cilium/                 # Example: Cilium implementation
│   ├── installer.go
│   └── installer_test.go
└── yourcni/               # Your new CNI goes here
    ├── installer.go
    └── installer_test.go
```

## Step-by-Step Guide

### 1. Create CNI Package Directory

```bash
# From repository root
mkdir -p pkg/svc/installer/cni/yourcni
cd pkg/svc/installer/cni/yourcni
```

### 2. Implement Installer

Create `installer.go`:

```go
package yourcniinstaller

import (
    "context"
    "fmt"
    "time"

    "github.com/devantler-tech/ksail-go/pkg/client/helm"
    "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"
    "github.com/devantler-tech/ksail-go/pkg/svc/installer/k8sutil"
)

// YourCNIInstaller implements the installer.Installer interface.
type YourCNIInstaller struct {
    *cni.CNIInstallerBase
}

// NewYourCNIInstaller creates a new installer instance.
func NewYourCNIInstaller(
    client helm.Interface,
    kubeconfig, context string,
    timeout time.Duration,
) *YourCNIInstaller {
    installer := &YourCNIInstaller{}
    installer.CNIInstallerBase = cni.NewCNIInstallerBase(
        client,
        kubeconfig,
        context,
        timeout,
        installer.waitForReadiness,
    )
    return installer
}

// Install installs your CNI via Helm.
func (y *YourCNIInstaller) Install(ctx context.Context) error {
    err := y.helmInstallOrUpgradeYourCNI(ctx)
    if err != nil {
        return fmt.Errorf("failed to install YourCNI: %w", err)
    }
    return nil
}

// SetWaitForReadinessFunc overrides the readiness wait function (for testing).
func (y *YourCNIInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
    y.CNIInstallerBase.SetWaitForReadinessFunc(waitFunc, y.waitForReadiness)
}

// Uninstall removes the Helm release.
func (y *YourCNIInstaller) Uninstall(ctx context.Context) error {
    client, err := y.GetClient()
    if err != nil {
        return fmt.Errorf("get helm client: %w", err)
    }

    err = client.UninstallRelease(ctx, "yourcni", "kube-system")
    if err != nil {
        return fmt.Errorf("failed to uninstall yourcni release: %w", err)
    }
    return nil
}

// --- Internal methods ---

func (y *YourCNIInstaller) helmInstallOrUpgradeYourCNI(ctx context.Context) error {
    client, err := y.GetClient()
    if err != nil {
        return fmt.Errorf("get helm client: %w", err)
    }

    // Configure Helm repository
    // Note: Name and RepoName serve different purposes - see pkg/svc/installer/cni_helpers.go for details
    repoConfig := cni.HelmRepoConfig{
        Name:     "yourcni",                   // Repository identifier used in Helm commands (e.g., "helm repo add <Name> <URL>")
        URL:      "https://helm.yourcni.io",  // Update with actual repo URL
        RepoName: "yourcni",                   // Human-readable name used in error messages (can differ from Name, see Calico installer example)
    }

    // Configure Helm chart installation
    // Note: RepoURL must match the URL field in repoConfig
    chartConfig := cni.HelmChartConfig{
        ReleaseName:     "yourcni",
        ChartName:       "yourcni/yourcni",               // Format: repo-name/chart-name
        Namespace:       "kube-system",
        RepoURL:         "https://helm.yourcni.io",      // Must match repoConfig.URL
        CreateNamespace: false,
        SetJSONVals:     defaultYourCNIValues(),
    }

    err = cni.InstallOrUpgradeHelmChart(ctx, client, repoConfig, chartConfig, y.GetTimeout())
    if err != nil {
        return fmt.Errorf("install or upgrade yourcni: %w", err)
    }
    return nil
}

func defaultYourCNIValues() map[string]string {
    // Add your CNI's default Helm values here
    return map[string]string{
        "key": "value",
    }
}

func (y *YourCNIInstaller) waitForReadiness(ctx context.Context) error {
    checks := []k8sutil.ReadinessCheck{
        {Type: "daemonset", Namespace: "kube-system", Name: "yourcni"},
        // Add more checks as needed (deployments, statefulsets, etc.)
    }

    err := cni.WaitForResourceReadiness(
        ctx,
        y.GetKubeconfig(),
        y.GetContext(),
        checks,
        y.GetTimeout(),
        "yourcni",
    )
    if err != nil {
        return fmt.Errorf("wait for yourcni readiness: %w", err)
    }
    return nil
}
```

### 3. Add Unit Tests

Create `installer_test.go`:

```go
package yourcniinstaller

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/devantler-tech/ksail-go/pkg/client/helm/mocks"
)

func TestYourCNIInstaller_Install(t *testing.T) {
    mockClient := mocks.NewMockInterface(t)
    installer := NewYourCNIInstaller(mockClient, "/path/to/kubeconfig", "test-context", 5*time.Minute)

    // Mock expected Helm operations
    mockClient.EXPECT().
        AddRepo(context.Background(), "yourcni", "https://helm.yourcni.io").
        Return(nil)

    mockClient.EXPECT().
        InstallOrUpgradeChart(
            context.Background(),
            "yourcni",
            "yourcni/yourcni",
            "kube-system",
            map[string]string{"key": "value"},
        ).
        Return(nil)

    // Override readiness check for testing
    installer.SetWaitForReadinessFunc(func(ctx context.Context) error {
        return nil  // Skip actual readiness check in unit tests
    })

    err := installer.Install(context.Background())
    assert.NoError(t, err)
}

func TestYourCNIInstaller_Uninstall(t *testing.T) {
    mockClient := mocks.NewMockInterface(t)
    installer := NewYourCNIInstaller(mockClient, "/path/to/kubeconfig", "test-context", 5*time.Minute)

    mockClient.EXPECT().
        UninstallRelease(context.Background(), "yourcni", "kube-system").
        Return(nil)

    err := installer.Uninstall(context.Background())
    assert.NoError(t, err)
}
```

### 4. Run Tests Locally

```bash
# Run your CNI's tests
go test ./pkg/svc/installer/cni/yourcni/...

# Run all CNI tests
go test ./pkg/svc/installer/cni/...

# Run with verbose output
go test -v ./pkg/svc/installer/cni/yourcni/...
```

### 5. Generate Mocks (if needed)

If you added new interfaces:

```bash
# From repository root
mockery
```

### 6. Validate Build

```bash
# Ensure everything compiles
go build ./...

# Run linter
golangci-lint run
```

### 7. Update Documentation

Add your CNI to `CONTRIBUTING.md`:

```markdown
## Supported CNIs

- Cilium (default)
- Calico
- YourCNI (added in vX.Y.Z)
```

### 8. Integration Test (Optional but Recommended)

Test your CNI with a real cluster:

```bash
# Update ksail.yaml to use your CNI
vim ksail.yaml  # Set cni: YourCNI

# Create cluster
ksail cluster up

# Verify CNI installation
kubectl get pods -n kube-system | grep yourcni

# Clean up
ksail cluster down
```

## Key Patterns to Follow

### 1. Always Embed CNIInstallerBase

```go
type YourCNIInstaller struct {
    *cni.CNIInstallerBase  // ✅ Always embed this
}
```

### 2. Use Shared Helper Functions

```go
// ✅ Good: Use shared helper
err := cni.InstallOrUpgradeHelmChart(ctx, client, repoConfig, chartConfig, timeout)

// ❌ Bad: Duplicate Helm installation logic
err := client.AddRepo(...)
err = client.InstallChart(...)
```

### 3. Implement Readiness Checks

```go
// ✅ Good: Check actual resources
checks := []k8sutil.ReadinessCheck{
    {Type: "daemonset", Namespace: "kube-system", Name: "yourcni"},
}

// ❌ Bad: No readiness verification
return nil  // Don't skip readiness checks
```

### 4. Follow Error Wrapping Convention

```go
// ✅ Good: Wrap with context
return fmt.Errorf("failed to install YourCNI: %w", err)

// ❌ Bad: Lose error context
return err
```

## Troubleshooting

### Import Errors

**Problem**: `cannot find package "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"`

**Solution**:

```bash
go mod tidy
go build ./...
```

### Test Failures

**Problem**: Tests fail with "readiness timeout"

**Solution**: Override readiness function in tests:

```go
installer.SetWaitForReadinessFunc(func(ctx context.Context) error {
    return nil  // Skip actual checks in unit tests
})
```

### Mock Generation Errors

**Problem**: `mockery` can't find interfaces

**Solution**: Ensure `.mockery.yml` includes your package:

```yaml
packages:
  github.com/devantler-tech/ksail-go/pkg/svc/installer/cni:
    interfaces:
      YourInterface:
        config:
          mockname: MockYourInterface
```

## Examples

See existing CNI implementations for reference:

- **Cilium**: `pkg/svc/installer/cni/cilium/installer.go` (comprehensive example)
- **Calico**: `pkg/svc/installer/cni/calico/installer.go` (alternative pattern)

## Getting Help

- Review existing CNI implementations in `pkg/svc/installer/cni/`
- Check `pkg/svc/installer/cni/base.go` for available helper methods
- Run `go doc github.com/devantler-tech/ksail-go/pkg/svc/installer/cni` for API docs
- Open an issue in the repository for questions

## Checklist

Before submitting your CNI implementation:

- [ ] Package created under `pkg/svc/installer/cni/yourcni/`
- [ ] `installer.go` implements `Install()` and `Uninstall()` methods
- [ ] Embeds `CNIInstallerBase` for shared functionality
- [ ] Uses `cni.InstallOrUpgradeHelmChart()` helper
- [ ] Implements `waitForReadiness()` with appropriate checks
- [ ] `installer_test.go` covers Install/Uninstall scenarios
- [ ] Tests pass: `go test ./pkg/svc/installer/cni/yourcni/...`
- [ ] Build succeeds: `go build ./...`
- [ ] Linter passes: `golangci-lint run`
- [ ] Documentation updated (CONTRIBUTING.md)
- [ ] Integration tested with real cluster (optional but recommended)
