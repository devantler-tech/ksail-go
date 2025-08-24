// Package k3dprovisioner contains the K3d cluster provisioner and its client provider interfaces.
package k3dprovisioner

import (
	"context"

	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
)

// K3dConfigProvider describes the subset of methods from k3d's config used here.
type K3dConfigProvider interface {
	TransformSimpleToClusterConfig(
		ctx context.Context,
		runtime runtimes.Runtime,
		simpleConfig v1alpha5.SimpleConfig,
		filename string,
	) (*v1alpha5.ClusterConfig, error)
}
