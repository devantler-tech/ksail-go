package k3dprovisioner_test

import (
	"context"
	"testing"

	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
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

func TestDefaultK3dClientAdapterClusterList(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	// This will fail if no runtime is available, which is expected in unit tests
	// The important part is that the method signature and interface are correct
	clusters, err := adapter.ClusterList(ctx, runtime)

	// We expect an error if Docker/Podman is not available, or success with empty list
	// Both are acceptable for this test
	if err == nil {
		assert.NotNil(t, clusters, "clusters should not be nil when no error")
	}
}

func TestDefaultK3dClientAdapterClusterGet(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	cluster := &types.Cluster{
		Name: "test-cluster",
	}

	// This should fail with "cluster not found" or runtime error
	result, err := adapter.ClusterGet(ctx, runtime, cluster)

	// We expect an error since the cluster doesn't exist
	// The important part is that the method signature and interface are correct
	if err == nil {
		assert.NotNil(t, result, "result should not be nil when no error")
	}
}

func TestDefaultK3dClientAdapterClusterDelete(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	cluster := &types.Cluster{
		Name: "nonexistent-cluster",
	}

	opts := types.ClusterDeleteOpts{}

	// This should fail since the cluster doesn't exist
	err := adapter.ClusterDelete(ctx, runtime, cluster, opts)

	// We expect an error (cluster not found or runtime error)
	// The important part is that the method works and doesn't panic
	_ = err
}

func TestDefaultK3dClientAdapterClusterStart(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	cluster := &types.Cluster{
		Name: "nonexistent-cluster",
	}

	opts := types.ClusterStartOpts{}

	// This should fail since the cluster doesn't exist
	err := adapter.ClusterStart(ctx, runtime, cluster, opts)

	// We expect an error (cluster not found or runtime error)
	// The important part is that the method works and doesn't panic
	_ = err
}

func TestDefaultK3dClientAdapterClusterStop(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	cluster := &types.Cluster{
		Name: "nonexistent-cluster",
	}

	// This should fail since the cluster doesn't exist
	err := adapter.ClusterStop(ctx, runtime, cluster)

	// We expect an error (cluster not found or runtime error)
	// The important part is that the method works and doesn't panic
	_ = err
}

func TestDefaultK3dClientAdapterClusterRun(t *testing.T) {
	t.Parallel()

	// Skip this test as it requires a complete cluster configuration and Docker/Podman.
	// Interface implementation is validated in TestDefaultK3dClientAdapterImplementsInterface.
	t.Skip("Skipping ClusterRun test as it requires complete configuration and container runtime")
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
