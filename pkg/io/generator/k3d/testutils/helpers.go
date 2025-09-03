// Package testutils provides test utilities specific to K3d generator functionality.
package testutils

import (
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1/testutils"
)

// CreateDefaultK3dSpec creates a default v1alpha1.Spec configured for K3d for testing.
func CreateDefaultK3dSpec() v1alpha1.Spec {
	spec := clustertestutils.CreateDefaultSpec()
	spec.Distribution = v1alpha1.DistributionK3d

	return spec
}