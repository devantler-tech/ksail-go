package stubs

import (
	"context"

	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
)

// K3dConfigProviderStub is a stub implementation of K3dConfigProvider interface.
type K3dConfigProviderStub struct {
	TransformResult *v1alpha5.ClusterConfig
	TransformError  error

	TransformCalls int
}

// NewK3dConfigProviderStub creates a new K3dConfigProviderStub.
func NewK3dConfigProviderStub() *K3dConfigProviderStub {
	return &K3dConfigProviderStub{
		TransformResult: &v1alpha5.ClusterConfig{
			// Create minimal config structure for stub
		},
	}
}

// TransformSimpleToClusterConfig simulates config transformation.
func (k *K3dConfigProviderStub) TransformSimpleToClusterConfig(
	ctx context.Context,
	runtime runtimes.Runtime,
	simpleConfig v1alpha5.SimpleConfig,
	filename string,
) (*v1alpha5.ClusterConfig, error) {
	k.TransformCalls++
	return k.TransformResult, k.TransformError
}
