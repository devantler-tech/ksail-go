package k3dprovisioner_test

import (
	"testing"

	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestDefaultK3dClientAdapterUsageInProvisioner(t *testing.T) {
	t.Parallel()

	provisioner := createK3dProvisionerWithDefaultAdapters(t)

	require.NotNil(t, provisioner, "provisioner should not be nil")
}

// createK3dProvisionerWithDefaultAdapters is a helper function to create a K3d provisioner
// with default adapters, reducing code duplication across tests.
func createK3dProvisionerWithDefaultAdapters(
	t *testing.T,
) *k3dprovisioner.K3dClusterProvisioner {
	t.Helper()

	simpleCfg := &v1alpha5.SimpleConfig{}
	simpleCfg.Name = "test-cluster-" + t.Name()

	clientAdapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	configAdapter := k3dprovisioner.NewDefaultK3dConfigAdapter()

	return k3dprovisioner.NewK3dClusterProvisioner(
		simpleCfg,
		clientAdapter,
		configAdapter,
	)
}
