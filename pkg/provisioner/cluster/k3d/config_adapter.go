package k3dprovisioner

import (
	"context"

	k3dconfig "github.com/k3d-io/k3d/v5/pkg/config"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
)

// K3dConfigAdapter wraps k3d config functions to implement K3dConfigProvider interface
type K3dConfigAdapter struct{}

// NewK3dConfigAdapter creates a new K3d config adapter
func NewK3dConfigAdapter() *K3dConfigAdapter {
	return &K3dConfigAdapter{}
}

// TransformSimpleToClusterConfig transforms a simple config to a full cluster config
func (a *K3dConfigAdapter) TransformSimpleToClusterConfig(
	ctx context.Context,
	runtime runtimes.Runtime,
	simpleConfig v1alpha5.SimpleConfig,
	filename string,
) (*v1alpha5.ClusterConfig, error) {
	return k3dconfig.TransformSimpleToClusterConfig(ctx, runtime, simpleConfig, filename)
}
