package k3dprovisioner

import (
	"context"
	"fmt"

	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClientAdapter wraps k3d client functions to implement K3dClientProvider interface.
type K3dClientAdapter struct{}

// NewK3dClientAdapter creates a new K3d client adapter.
func NewK3dClientAdapter() *K3dClientAdapter {
	return &K3dClientAdapter{}
}

// ClusterRun creates and starts a k3d cluster.
func (a *K3dClientAdapter) ClusterRun(
	ctx context.Context,
	runtime runtimes.Runtime,
	clusterConfig *v1alpha5.ClusterConfig,
) error {
	err := k3dclient.ClusterRun(ctx, runtime, clusterConfig)
	if err != nil {
		return fmt.Errorf("failed to run k3d cluster %s: %w", clusterConfig.Name, err)
	}

	return nil
}

// ClusterDelete deletes a k3d cluster.
func (a *K3dClientAdapter) ClusterDelete(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterDeleteOpts,
) error {
	err := k3dclient.ClusterDelete(ctx, runtime, cluster, opts)
	if err != nil {
		return fmt.Errorf("failed to delete k3d cluster %s: %w", cluster.Name, err)
	}

	return nil
}

// ClusterGet retrieves information about a k3d cluster.
func (a *K3dClientAdapter) ClusterGet(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) (*types.Cluster, error) {
	res, err := k3dclient.ClusterGet(ctx, runtime, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get k3d cluster %s: %w", cluster.Name, err)
	}

	return res, nil
}

// ClusterStart starts an existing k3d cluster.
func (a *K3dClientAdapter) ClusterStart(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterStartOpts,
) error {
	err := k3dclient.ClusterStart(ctx, runtime, cluster, opts)
	if err != nil {
		return fmt.Errorf("failed to start k3d cluster %s: %w", cluster.Name, err)
	}

	return nil
}

// ClusterStop stops a running k3d cluster.
func (a *K3dClientAdapter) ClusterStop(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) error {
	err := k3dclient.ClusterStop(ctx, runtime, cluster)
	if err != nil {
		return fmt.Errorf("failed to stop k3d cluster %s: %w", cluster.Name, err)
	}

	return nil
}

// ClusterList lists all k3d clusters.
func (a *K3dClientAdapter) ClusterList(
	ctx context.Context,
	runtime runtimes.Runtime,
) ([]*types.Cluster, error) {
	res, err := k3dclient.ClusterList(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to list k3d clusters: %w", err)
	}

	return res, nil
}
