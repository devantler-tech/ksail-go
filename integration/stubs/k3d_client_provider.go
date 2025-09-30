package stubs

import (
	"context"

	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClientProviderStub is a stub implementation of K3dClientProvider interface.
type K3dClientProviderStub struct {
	ClusterRunError    error
	ClusterDeleteError error
	ClusterGetResult   *types.Cluster
	ClusterGetError    error
	ClusterStartError  error
	ClusterStopError   error
	ClusterListResult  []*types.Cluster
	ClusterListError   error

	RunCalls    int
	DeleteCalls int
	GetCalls    int
	StartCalls  int
	StopCalls   int
	ListCalls   int
}

// NewK3dClientProviderStub creates a new K3dClientProviderStub.
func NewK3dClientProviderStub() *K3dClientProviderStub {
	return &K3dClientProviderStub{
		ClusterGetResult: &types.Cluster{Name: "test-cluster"},
		ClusterListResult: []*types.Cluster{
			{Name: "k3d-cluster"},
		},
	}
}

// ClusterRun simulates cluster creation.
func (k *K3dClientProviderStub) ClusterRun(
	ctx context.Context,
	runtime runtimes.Runtime,
	clusterConfig *v1alpha5.ClusterConfig,
) error {
	k.RunCalls++
	return k.ClusterRunError
}

// ClusterDelete simulates cluster deletion.
func (k *K3dClientProviderStub) ClusterDelete(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterDeleteOpts,
) error {
	k.DeleteCalls++
	return k.ClusterDeleteError
}

// ClusterGet simulates cluster retrieval.
func (k *K3dClientProviderStub) ClusterGet(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) (*types.Cluster, error) {
	k.GetCalls++
	return k.ClusterGetResult, k.ClusterGetError
}

// ClusterStart simulates cluster start.
func (k *K3dClientProviderStub) ClusterStart(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
	opts types.ClusterStartOpts,
) error {
	k.StartCalls++
	return k.ClusterStartError
}

// ClusterStop simulates cluster stop.
func (k *K3dClientProviderStub) ClusterStop(
	ctx context.Context,
	runtime runtimes.Runtime,
	cluster *types.Cluster,
) error {
	k.StopCalls++
	return k.ClusterStopError
}

// ClusterList simulates cluster listing.
func (k *K3dClientProviderStub) ClusterList(
	ctx context.Context,
	runtime runtimes.Runtime,
) ([]*types.Cluster, error) {
	k.ListCalls++
	return k.ClusterListResult, k.ClusterListError
}
