# Quickstart: Flannel CNI Implementation

**Date**: 2025-11-15
**Estimated Duration**: 2-3 hours for core implementation
**Prerequisite**: Read [research.md](./research.md), [data-model.md](./data-model.md), and [contracts/](./contracts/)

## TL;DR

Add Flannel CNI support to KSail-Go by:

1. Add `CNIFlannel` constant to `pkg/apis/cluster/v1alpha1/types.go`
2. Create `pkg/client/kubectl` package for manifest application
3. Create `pkg/svc/installer/cni/flannel/` with installer implementation
4. Update cluster create command to handle Flannel
5. Add tests following existing patterns

**Pattern**: Mirror Cilium installer but use kubectl manifest instead of Helm

## Prerequisites

- Go 1.25.4+ installed
- Repository cloned and dependencies downloaded (`go mod download`)
- Familiarity with existing CNI installer pattern (see `pkg/svc/installer/cni/cilium/`)
- mockery v3.x installed for mock generation

## Step-by-Step Implementation

### Phase 1: Add CNI Enum Value (15 minutes)

**File**: `pkg/apis/cluster/v1alpha1/types.go`

1. Add constant after CNICalico:

```go
const (
    CNIDefault CNI = "Default"
    CNICilium  CNI = "Cilium"
    CNICalico  CNI = "Calico"
    CNIFlannel CNI = "Flannel"  // ADD THIS
)
```

1. Update `validCNIs()` function:

```go
func validCNIs() []CNI {
    return []CNI{CNIDefault, CNICilium, CNICalico, CNIFlannel}  // Add CNIFlannel
}
```

1. Update `Set()` method switch:

```go
func (c *CNI) Set(value string) error {
    switch CNI(value) {
    case CNIDefault, CNICilium, CNICalico, CNIFlannel:  // Add CNIFlannel
        *c = CNI(value)
        return nil
    default:
        return fmt.Errorf("%w: %s (valid: %s, %s, %s, %s)",
            ErrInvalidCNI, value, CNIDefault, CNICilium, CNICalico, CNIFlannel)
    }
}
```

**Verify**: Run `go build ./pkg/apis/cluster/...`

### Phase 2: Add Validation Tests (20 minutes)

**File**: `pkg/apis/cluster/v1alpha1/types_test.go`

Add test cases to existing `TestCNI_Set`:

```go
{
    name:    "valid Flannel",
    input:   "Flannel",
    want:    CNIFlannel,
    wantErr: false,
},
{
    name:    "invalid flannel lowercase",
    input:   "flannel",
    want:    "",
    wantErr: true,
},
```

**Verify**: `go test ./pkg/apis/cluster/v1alpha1/...`

### Phase 3: Create kubectl Client Package (30 minutes)

**New Package**: `pkg/client/kubectl/`

1. Create `pkg/client/kubectl/client.go`:

```go
package kubectl

import (
    "context"
    "fmt"
    "io"
    "net/http"

    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/util/yaml"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
)

//go:generate mockery --name=Interface --output=. --filename=mocks.go --outpkg=kubectl

// Interface defines kubectl operations
type Interface interface {
    Apply(ctx context.Context, manifestURL string) error
}

// Client implements kubectl operations using Kubernetes dynamic client
type Client struct {
    dynamicClient dynamic.Interface
    restConfig    *rest.Config
}

// New creates a new kubectl client
func New(kubeconfig, context string) (Interface, error) {
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, fmt.Errorf("build config: %w", err)
    }

    dynamicClient, err := dynamic.NewForConfig(config)
    if err != nil {
        return nil, fmt.Errorf("create dynamic client: %w", err)
    }

    return &Client{
        dynamicClient: dynamicClient,
        restConfig:    config,
    }, nil
}

// Apply fetches and applies a manifest from URL
func (c *Client) Apply(ctx context.Context, manifestURL string) error {
    // Fetch manifest
    resp, err := http.Get(manifestURL)
    if err != nil {
        return fmt.Errorf("fetch manifest: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("fetch manifest: unexpected status %d", resp.StatusCode)
    }

    // Parse and apply resources
    decoder := yaml.NewYAMLOrJSONDecoder(resp.Body, 4096)
    for {
        var obj unstructured.Unstructured
        err := decoder.Decode(&obj)
        if err == io.EOF {
            break
        }
        if err != nil {
            return fmt.Errorf("decode manifest: %w", err)
        }

        // Apply object (simplified - production needs GVR mapping)
        // TODO: Implement proper GVR mapping and apply logic
    }

    return nil
}
```

1. Create `pkg/client/kubectl/doc.go`:

```go
// Package kubectl provides utilities for applying Kubernetes manifests
package kubectl
```

1. Generate mocks: `mockery --name=Interface --dir=pkg/client/kubectl`

**Verify**: `go build ./pkg/client/kubectl/...`

### Phase 4: Create Flannel Installer (45 minutes)

**New Package**: `pkg/svc/installer/cni/flannel/`

1. Create `pkg/svc/installer/cni/flannel/installer.go`:

```go
package flannelinstaller

import (
    "context"
    "fmt"
    "time"

    "github.com/devantler-tech/ksail-go/pkg/client/kubectl"
    "github.com/devantler-tech/ksail-go/pkg/k8s"
    "github.com/devantler-tech/ksail-go/pkg/svc/installer"
    "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni"
)

const (
    flannelManifestURL = "https://github.com/flannel-io/flannel/releases/latest/download/kube-flannel.yml"
)

// FlannelInstaller implements the installer.Installer interface for Flannel
type FlannelInstaller struct {
    *cni.InstallerBase
    kubectlClient kubectl.Interface
}

// NewFlannelInstaller creates a new Flannel installer instance
func NewFlannelInstaller(
    client kubectl.Interface,
    kubeconfig, context string,
    timeout time.Duration,
) *FlannelInstaller {
    flannelInstaller := &FlannelInstaller{
        kubectlClient: client,
    }
    flannelInstaller.InstallerBase = cni.NewInstallerBase(
        nil, // No Helm client needed
        kubeconfig,
        context,
        timeout,
        flannelInstaller.waitForReadiness,
    )

    return flannelInstaller
}

// Install applies the Flannel manifest and waits for readiness
func (f *FlannelInstaller) Install(ctx context.Context) error {
    err := f.kubectlClient.Apply(ctx, flannelManifestURL)
    if err != nil {
        return fmt.Errorf("apply flannel manifest: %w", err)
    }

    return nil
}

// Uninstall removes Flannel components (simplified)
func (f *FlannelInstaller) Uninstall(ctx context.Context) error {
    // TODO: Implement manifest deletion
    return fmt.Errorf("uninstall not yet implemented")
}

// SetWaitForReadinessFunc overrides readiness wait (for testing)
func (f *FlannelInstaller) SetWaitForReadinessFunc(waitFunc func(context.Context) error) {
    f.InstallerBase.SetWaitForReadinessFunc(waitFunc, f.waitForReadiness)
}

func (f *FlannelInstaller) waitForReadiness(ctx context.Context) error {
    checks := []k8s.ReadinessCheck{
        {Type: "daemonset", Namespace: "kube-flannel", Name: "kube-flannel-ds"},
    }

    err := installer.WaitForResourceReadiness(
        ctx,
        f.GetKubeconfig(),
        f.GetContext(),
        checks,
        f.GetTimeout(),
        "flannel",
    )
    if err != nil {
        return fmt.Errorf("wait for flannel readiness: %w", err)
    }

    return nil
}
```

1. Create `pkg/svc/installer/cni/flannel/doc.go`:

```go
// Package flannelinstaller provides Flannel CNI installation
package flannelinstaller
```

**Verify**: `go build ./pkg/svc/installer/cni/flannel/...`

### Phase 5: Add Installer Tests (45 minutes)

**File**: `pkg/svc/installer/cni/flannel/installer_test.go`

```go
package flannelinstaller_test

import (
    "context"
    "testing"
    "time"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "github.com/devantler-tech/ksail-go/pkg/client/kubectl"
    flannelinstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cni/flannel"
)

func TestFlannelInstaller_Install(t *testing.T) {
    tests := []struct {
        name          string
        mockSetup     func(*kubectl.MockInterface)
        wantErr       bool
        errorContains string
    }{
        {
            name: "successful installation",
            mockSetup: func(m *kubectl.MockInterface) {
                m.On("Apply", mock.Anything, mock.Anything).Return(nil)
            },
            wantErr: false,
        },
        {
            name: "network error",
            mockSetup: func(m *kubectl.MockInterface) {
                m.On("Apply", mock.Anything, mock.Anything).
                    Return(fmt.Errorf("network unavailable"))
            },
            wantErr:       true,
            errorContains: "network unavailable",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockClient := new(kubectl.MockInterface)
            tt.mockSetup(mockClient)

            installer := flannelinstaller.NewFlannelInstaller(
                mockClient,
                "/fake/kubeconfig",
                "fake-context",
                5*time.Minute,
            )

            // Skip readiness check for unit tests
            installer.SetWaitForReadinessFunc(func(ctx context.Context) error {
                return nil
            })

            err := installer.Install(context.Background())

            if tt.wantErr {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorContains)
            } else {
                assert.NoError(t, err)
            }

            mockClient.AssertExpectations(t)
        })
    }
}
```

**Verify**: `go test ./pkg/svc/installer/cni/flannel/...`

### Phase 6: Integrate with Cluster Create Command (20 minutes)

**File**: `cmd/cluster/create.go`

Find the CNI installer factory switch and add Flannel case:

```go
switch config.Spec.CNI {
case v1alpha1.CNIDefault:
    // existing...
case v1alpha1.CNICilium:
    // existing...
case v1alpha1.CNICalico:
    // existing...
case v1alpha1.CNIFlannel:
    kubectlClient, err := kubectl.New(config.Spec.Connection.Kubeconfig, config.Spec.Connection.Context)
    if err != nil {
        return fmt.Errorf("create kubectl client: %w", err)
    }
    cniInstaller = flannelinstaller.NewFlannelInstaller(
        kubectlClient,
        config.Spec.Connection.Kubeconfig,
        config.Spec.Connection.Context,
        timeout,
    )
default:
    return fmt.Errorf("unsupported CNI: %s", config.Spec.CNI)
}
```

**Verify**: `go build ./cmd/cluster/...`

### Phase 7: Update Schema (5 minutes)

Regenerate JSON schema:

```bash
go run .github/scripts/generate-schema.go
```

**Verify**: `git diff schemas/ksail-config.schema.json` should show Flannel in enum

### Phase 8: System Tests (30 minutes)

**File**: `.github/workflows/ci.yaml`

Add Flannel to system test matrix (if exists):

```yaml
matrix:
  distribution: [Kind, K3d]
  cni: [Default, Cilium, Calico, Flannel]  # Add Flannel
```

Run local test:

```bash
# Create test project
mkdir /tmp/flannel-test && cd /tmp/flannel-test
ksail cluster init --distribution Kind --cni Flannel

# Verify ksail.yaml
cat ksail.yaml | grep "cni: Flannel"

# Create cluster (requires Docker)
ksail up

# Verify Flannel running
kubectl get daemonset -n kube-flannel

# Cleanup
ksail down
```

## Verification Checklist

- [ ] `go build ./...` succeeds
- [ ] `go test ./...` passes
- [ ] `mockery` generates all mocks
- [ ] `golangci-lint run` passes
- [ ] Schema includes Flannel in CNI enum
- [ ] `ksail cluster init --cni Flannel` works
- [ ] Generated ksail.yaml has `cni: Flannel`
- [ ] Cluster creation succeeds (manual test)
- [ ] Flannel DaemonSet is running
- [ ] Nodes reach Ready state
- [ ] Pod-to-pod communication works

## Common Pitfalls

1. **Case sensitivity**: Use "Flannel" not "flannel"
2. **Namespace**: Flannel uses `kube-flannel` not `kube-system`
3. **Resource name**: DaemonSet is `kube-flannel-ds` not `flannel`
4. **kubectl client**: Must implement proper GVR mapping for production
5. **InstallerBase**: Must pass `nil` for Helm client (Flannel doesn't use Helm)

## Next Steps

After core implementation:

1. Add documentation to `docs/cni.md`
2. Update README.md CNI options list
3. Add E2E networking tests
4. Implement proper error diagnostics
5. Add rollback handling per FR-011a

## Estimated Timeline

- Phase 1-2: 35 minutes (Enum + tests)
- Phase 3: 30 minutes (kubectl package)
- Phase 4-5: 90 minutes (Installer + tests)
- Phase 6-7: 25 minutes (Integration + schema)
- Phase 8: 30 minutes (System tests)

**Total**: ~3.5 hours for full implementation and testing

## Reference Implementation

See existing CNI installers for patterns:

- `pkg/svc/installer/cni/cilium/installer.go` - Overall structure
- `pkg/svc/installer/cni/cilium/installer_test.go` - Test patterns
- `pkg/client/helm` - Client abstraction pattern (mirror for kubectl)
