package k3dprovisioner_test

import (
	"testing"

	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	"github.com/stretchr/testify/assert"
)

func TestNewDefaultK3dClientAdapter(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dClientAdapter()

	assert.NotNil(t, adapter, "adapter should not be nil")
}

func TestDefaultK3dClientAdapterImplementsInterface(t *testing.T) {
	t.Parallel()

	// Verify that DefaultK3dClientAdapter implements K3dClientProvider interface
	var _ k3dprovisioner.K3dClientProvider = (*k3dprovisioner.DefaultK3dClientAdapter)(nil)
}
