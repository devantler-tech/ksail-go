package k3dprovisioner_test

import (
	"testing"

	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultK3dConfigAdapter(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dConfigAdapter()

	assert.NotNil(t, adapter, "adapter should not be nil")
}

func TestDefaultK3dConfigAdapterImplementsInterface(t *testing.T) {
	t.Parallel()

	// Verify that DefaultK3dConfigAdapter implements K3dConfigProvider interface
	var _ k3dprovisioner.K3dConfigProvider = (*k3dprovisioner.DefaultK3dConfigAdapter)(nil)
}
