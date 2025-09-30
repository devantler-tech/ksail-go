package k3dprovisioner

import (
	"context"

	k3dclient "github.com/k3d-io/k3d/v5/pkg/client"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClientAdapter wraps k3d client functions to implement K3dClientProvider interface
type K3dClientAdapter struct{}

// NewK3dClientAdapter creates a new K3d client adapter
func NewK3dClientAdapter() *K3dClientAdapter {
	return &K3dClientAdapter{}
}

// ClusterRun creates and starts a k3d cluster
func (a *K3dClientAdapter) ClusterRun(
	ctx context.Context,
	runtime runtimes.Runtime,
	clusterConfig *v1alpha5.ClusterConfig,
) error {
	return k3dclient.ClusterRun(ctx, runtime, clusterConfig)
}

// ClusterDelete deletes a k3d cluster
func (a *K3dClientAdapter) ClusterDelete(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterDeleteOpts,
) error {
	return k3dclient.ClusterDelete(ctx, runtime, cluster, opts)
}

// ClusterGet retrieves information about a k3d cluster
func (a *K3dClientAdapter) ClusterGet(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) (*types.Cluster, error) {
	return k3dclient.ClusterGet(ctx, runtime, cluster)
}

// ClusterStart starts an existing k3d cluster
func (a *K3dClientAdapter) ClusterStart(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterStartOpts,
) error {
	return k3dclient.ClusterStart(ctx, runtime, cluster, opts)
}

// ClusterStop stops a running k3d cluster
func (a *K3dClientAdapter) ClusterStop(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) error {
	return k3dclient.ClusterStop(ctx, runtime, cluster)
}

// ClusterList lists all k3d clusters
func (a *K3dClientAdapter) ClusterList(
	ctx context.Context,
	runtime runtimes.Runtime,
) ([]*types.Cluster, error) {
	return k3dclient.ClusterList(ctx, runtime)
}
