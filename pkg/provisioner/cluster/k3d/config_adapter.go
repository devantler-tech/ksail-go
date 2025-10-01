package k3dprovisioner

import (
	"context"
	"fmt"

	"github.com/k3d-io/k3d/v5/pkg/config"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
)

// DefaultK3dConfigAdapter provides a production-ready implementation of K3dConfigProvider
// that wraps the k3d library's config functions.
type DefaultK3dConfigAdapter struct{}

// NewDefaultK3dConfigAdapter creates a new instance of the default K3d config adapter.
func NewDefaultK3dConfigAdapter() *DefaultK3dConfigAdapter {
	return &DefaultK3dConfigAdapter{}
}

// TransformSimpleToClusterConfig transforms a simple configuration into a full cluster configuration.
func (a *DefaultK3dConfigAdapter) TransformSimpleToClusterConfig(
	ctx context.Context,
	runtime runtimes.Runtime,
	simpleConfig v1alpha5.SimpleConfig,
	filename string,
) (*v1alpha5.ClusterConfig, error) {
	clusterConfig, err := config.TransformSimpleToClusterConfig(
		ctx,
		runtime,
		simpleConfig,
		filename,
	)
	if err != nil {
		return nil, fmt.Errorf("transform simple to cluster config: %w", err)
	}

	return clusterConfig, nil
}
