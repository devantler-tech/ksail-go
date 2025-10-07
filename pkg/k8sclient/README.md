# pkg/k8sclient

This package provides utilities for creating and managing Kubernetes clients for KSail.

## Purpose

Implements client creation and component status retrieval for interacting with Kubernetes clusters. This package provides abstractions over the `client-go` library to simplify Kubernetes API interactions.

## Features

- **Kubernetes Client Creation**: Creates clientsets from kubeconfig files, contexts, or in-cluster configuration
- **Component Status Retrieval**: Fetches component statuses from Kubernetes clusters (equivalent to `kubectl get cs`)
- **Flexible Configuration**: Supports custom kubeconfig paths, contexts, and automatic fallback to in-cluster config

## Usage

### Creating a Kubernetes Client

```go
import "github.com/devantler-tech/ksail-go/pkg/k8sclient"

// Create a client provider
clientProvider := k8sclient.NewDefaultClientProvider()

// Create a clientset with default kubeconfig (~/.kube/config)
clientset, err := clientProvider.CreateClient("", "")
if err != nil {
    log.Fatal("Failed to create client:", err)
}

// Create a clientset with custom kubeconfig and context
clientset, err := clientProvider.CreateClient("/path/to/kubeconfig", "my-context")
if err != nil {
    log.Fatal("Failed to create client:", err)
}
```

### Fetching Component Statuses

```go
import (
    "context"
    "github.com/devantler-tech/ksail-go/pkg/k8sclient"
)

// Create providers
clientProvider := k8sclient.NewDefaultClientProvider()
statusProvider := k8sclient.NewDefaultComponentStatusProvider()

// Create client
clientset, err := clientProvider.CreateClient("", "")
if err != nil {
    log.Fatal("Failed to create client:", err)
}

// Get component statuses
ctx := context.Background()
statuses, err := statusProvider.GetComponentStatuses(ctx, clientset)
if err != nil {
    log.Fatal("Failed to get component statuses:", err)
}

// Display statuses
for _, status := range statuses {
    fmt.Printf("Component: %s\n", status.Name)
    for _, condition := range status.Conditions {
        fmt.Printf("  %s: %s\n", condition.Type, condition.Status)
    }
}
```

## Interfaces

### ClientProvider

```go
type ClientProvider interface {
    CreateClient(kubeconfig, context string) (*kubernetes.Clientset, error)
}
```

Creates Kubernetes clientsets with support for:
- Custom kubeconfig paths
- Custom contexts
- Automatic fallback to default kubeconfig location (`~/.kube/config`)
- In-cluster configuration

### ComponentStatusProvider

```go
type ComponentStatusProvider interface {
    GetComponentStatuses(ctx context.Context, clientset *kubernetes.Clientset) ([]corev1.ComponentStatus, error)
}
```

Retrieves component statuses from a Kubernetes cluster, providing information about:
- Scheduler status
- Controller Manager status
- etcd status

---

[⬅️ Go Back](../README.md)
