package k3dprovisioner

import (
	"context"
	"fmt"

	"github.com/k3d-io/k3d/v5/pkg/client"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// DefaultK3dClientAdapter provides a production-ready implementation of K3dClientProvider
// that wraps the k3d library's client functions.
type DefaultK3dClientAdapter struct{}

// NewDefaultK3dClientAdapter creates a new instance of the default K3d client adapter.
func NewDefaultK3dClientAdapter() *DefaultK3dClientAdapter {
	return &DefaultK3dClientAdapter{}
}

// ClusterRun creates and starts a k3d cluster using the provided configuration.
func (a *DefaultK3dClientAdapter) ClusterRun(
	ctx context.Context,
	runtime runtimes.Runtime,
	clusterConfig *v1alpha5.ClusterConfig,
) error {
	err := client.ClusterRun(ctx, runtime, clusterConfig)
	if err != nil {
		return fmt.Errorf("cluster run: %w", err)
	}

	return nil
}

// ClusterDelete deletes a k3d cluster.
func (a *DefaultK3dClientAdapter) ClusterDelete(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterDeleteOpts,
) error {
	err := client.ClusterDelete(ctx, runtime, cluster, opts)
	if err != nil {
		return fmt.Errorf("cluster delete: %w", err)
	}

	return nil
}

// ClusterGet retrieves information about a k3d cluster.
func (a *DefaultK3dClientAdapter) ClusterGet(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) (*types.Cluster, error) {
	result, err := client.ClusterGet(ctx, runtime, cluster)
	if err != nil {
		return nil, fmt.Errorf("cluster get: %w", err)
	}

	return result, nil
}

// ClusterStart starts a stopped k3d cluster.
func (a *DefaultK3dClientAdapter) ClusterStart(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterStartOpts,
) error {
	err := client.ClusterStart(ctx, runtime, cluster, opts)
	if err != nil {
		return fmt.Errorf("cluster start: %w", err)
	}

	return nil
}

// ClusterStop stops a running k3d cluster.
func (a *DefaultK3dClientAdapter) ClusterStop(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) error {
	err := client.ClusterStop(ctx, runtime, cluster)
	if err != nil {
		return fmt.Errorf("cluster stop: %w", err)
	}

	return nil
}

// ClusterList lists all k3d clusters.
func (a *DefaultK3dClientAdapter) ClusterList(
	ctx context.Context,
	runtime runtimes.Runtime,
) ([]*types.Cluster, error) {
	clusters, err := client.ClusterList(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("cluster list: %w", err)
	}

	return clusters, nil
}

// KubeconfigGet retrieves the kubeconfig for a k3d cluster.
func (a *DefaultK3dClientAdapter) KubeconfigGet(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) (*clientcmdapi.Config, error) {
	kubeconfig, err := client.KubeconfigGet(ctx, runtime, cluster)
	if err != nil {
		return nil, fmt.Errorf("kubeconfig get: %w", err)
	}

	return kubeconfig, nil
}
