package k3dprovisioner

import (
	"context"

	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClientProvider describes the subset of methods from k3d's client used here.
type K3dClientProvider interface {
	ClusterRun(
		ctx context.Context,
		runtime runtimes.Runtime,
		clusterConfig *v1alpha5.ClusterConfig,
	) error
	ClusterDelete(
		ctx context.Context,
		runtime runtimes.Runtime,
		cluster *types.Cluster,
		opts types.ClusterDeleteOpts,
	) error
	ClusterGet(
		ctx context.Context,
		runtime runtimes.Runtime,
		cluster *types.Cluster,
	) (*types.Cluster, error)
	ClusterStart(
		ctx context.Context,
		runtime runtimes.Runtime,
		cluster *types.Cluster,
		opts types.ClusterStartOpts,
	) error
	ClusterStop(ctx context.Context, runtime runtimes.Runtime, cluster *types.Cluster) error
	ClusterList(ctx context.Context, runtime runtimes.Runtime) ([]*types.Cluster, error)
}
