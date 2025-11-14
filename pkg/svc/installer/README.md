# Installer Package

The `installer` package provides a unified interface for installing and uninstalling Kubernetes components and CNI plugins.

## Architecture

This package follows the **Single Responsibility Principle (SRP)** and **Separation of Concerns (SoC)** design principles. Each file has a focused responsibility:

### Core Files

- **`installer.go`** - Defines the `Installer` interface that all component installers implement
- **`readiness.go`** - High-level resource readiness checking orchestration
- **`cni_helpers.go`** - Base infrastructure for CNI installers

### Related Packages

- **`pkg/client/helm/config.go`** - Configuration structures for Helm repositories and charts
- **`pkg/client/helm/operations.go`** - Helm installation and upgrade operations
- **`pkg/k8s/`** - Kubernetes utilities (REST config, polling, resource readiness)

### Component Installers

Each component has its own subdirectory with focused responsibilities:

- **`calico/`** - Calico CNI installer
- **`cilium/`** - Cilium CNI installer
- **`argocd/`** - ArgoCD GitOps installer
- **`flux/`** - Flux GitOps installer
- **`istio/`** - Istio service mesh installer
- **`traefik/`** - Traefik ingress controller installer
- **`metrics-server/`** - Kubernetes metrics server installer
- **`applyset/`** - ApplySet resource installer

### Kubernetes Utilities (pkg/k8s)

The `pkg/k8s` package provides low-level Kubernetes operations, split into focused modules:

- **`rest_config.go`** - Kubernetes REST client configuration
- **`polling.go`** - Generic polling mechanism for readiness checks
- **`daemonset.go`** - DaemonSet-specific readiness checking
- **`deployment.go`** - Deployment-specific readiness checking
- **`multi_resource.go`** - Coordination of multiple resource readiness checks
- **`doc.go`** - Package documentation and overview

## Design Principles

### Single Responsibility Principle (SRP)

Each file and function has one clear responsibility:

- Configuration structures are separate from behavior
- Helm operations are isolated from Kubernetes operations
- Resource-specific logic (DaemonSet vs Deployment) is separated
- Generic polling logic is reusable across resource types

### Separation of Concerns (SoC)

The package is organized into distinct layers:

1. **Interface Layer** (`installer.go`) - Defines contracts
2. **Configuration Layer** (`config.go`) - Data structures only
3. **Operation Layer** (`helm_operations.go`, `readiness.go`) - Reusable operations
4. **Implementation Layer** (component subdirectories) - Specific installers
5. **Utility Layer** (`k8sutil/`) - Low-level Kubernetes operations

### Interface Segregation

- Small, focused interfaces (`Installer`)
- Components only depend on what they need
- Test mocking is straightforward

### Dependency Inversion

- High-level modules (`readiness.go`) depend on abstractions (`k8sutil` functions)
- Component installers depend on interfaces (`helm.Interface`)
- Concrete implementations are injected via constructors

## Usage Examples

### Installing a CNI Plugin

```go
import (
    "context"
    "time"
    
    "github.com/devantler-tech/ksail-go/pkg/client/helm"
    ciliuminstaller "github.com/devantler-tech/ksail-go/pkg/svc/installer/cilium"
)

func installCilium(kubeconfig, context string) error {
    helmClient := helm.NewHelmClient(kubeconfig, context)
    
    installer := ciliuminstaller.NewCiliumInstaller(
        helmClient,
        kubeconfig,
        context,
        5*time.Minute,
    )
    
    return installer.Install(context.Background())
}
```

### Using Helm Operations Directly

```go
import (
    "context"
    "time"
    
    "github.com/devantler-tech/ksail-go/pkg/client/helm"
)

func installCustomChart(client helm.Interface) error {
    repoConfig := helm.RepoConfig{
        Name:     "myrepo",
        URL:      "https://charts.example.com",
        RepoName: "Example Charts",
    }
    
    chartConfig := helm.ChartConfig{
        ReleaseName:     "myapp",
        ChartName:       "myrepo/myapp",
        Namespace:       "myapp",
        RepoURL:         "https://charts.example.com",
        CreateNamespace: true,
    }
    
    return helm.InstallOrUpgradeChart(
        context.Background(),
        client,
        repoConfig,
        chartConfig,
        5*time.Minute,
    )
}
```

### Checking Resource Readiness

```go
import (
    "context"
    "time"
    
    "github.com/devantler-tech/ksail-go/pkg/k8s"
    "github.com/devantler-tech/ksail-go/pkg/svc/installer"
)

func waitForComponents(kubeconfig, context string) error {
    checks := []k8s.ReadinessCheck{
        {Type: "deployment", Namespace: "kube-system", Name: "coredns"},
        {Type: "daemonset", Namespace: "kube-system", Name: "kube-proxy"},
    }
    
    return installer.WaitForResourceReadiness(
        context.Background(),
        kubeconfig,
        context,
        checks,
        5*time.Minute,
        "core components",
    )
}
```

## Testing

All components have comprehensive unit tests following the same patterns:

- Constructor validation
- Successful installation/uninstallation
- Error handling for each failure scenario
- Mock-based testing for external dependencies

Run tests with:

```bash
go test ./pkg/svc/installer/...
```

## Backward Compatibility

All public functions and types maintain their existing signatures. This refactoring only reorganizes code internally without breaking existing consumers.
