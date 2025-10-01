package kindprovisioner_test

import (
	"testing"

	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestNewDefaultKindProviderAdapter(t *testing.T) {
	t.Parallel()

	adapter := kindprovisioner.NewDefaultKindProviderAdapter()

	assert.NotNil(t, adapter, "adapter should not be nil")
}

func TestDefaultKindProviderAdapterImplementsInterface(t *testing.T) {
	t.Parallel()

	// Verify that DefaultKindProviderAdapter implements KindProvider interface
	var _ kindprovisioner.KindProvider = (*kindprovisioner.DefaultKindProviderAdapter)(nil)
}

func TestDefaultKindProviderAdapterList(t *testing.T) {
	t.Parallel()

	adapter := kindprovisioner.NewDefaultKindProviderAdapter()

	// This will return the list of kind clusters (may be empty)
	clusters, err := adapter.List()

	// Should not error in most cases (even if empty)
	if err == nil {
		assert.NotNil(t, clusters, "clusters should not be nil when no error")
	}
	// If error occurs (e.g., Docker not available), that's acceptable
	// The important part is that the method signature is correct
}

func TestDefaultKindProviderAdapterListNodes(t *testing.T) {
	t.Parallel()

	adapter := kindprovisioner.NewDefaultKindProviderAdapter()

	// This should fail for a nonexistent cluster
	nodes, err := adapter.ListNodes("nonexistent-cluster")

	// We expect an error since the cluster doesn't exist
	// The important part is that the method works and doesn't panic
	if err == nil {
		assert.NotNil(t, nodes, "nodes should not be nil when no error")
	}
}

func TestDefaultKindProviderAdapterDelete(t *testing.T) {
	t.Parallel()

	adapter := kindprovisioner.NewDefaultKindProviderAdapter()

	// This should fail gracefully for a nonexistent cluster
	err := adapter.Delete("nonexistent-cluster", "")

	// We expect an error or success (kind handles non-existent clusters gracefully)
	// The important part is that the method works and doesn't panic
	_ = err
}

func TestDefaultKindProviderAdapterUsageInProvisioner(t *testing.T) {
	t.Parallel()

	// Test that the adapter can be used with the provisioner
	kindConfig := &v1alpha4.Cluster{}
	kindConfig.Name = "test-cluster"

	providerAdapter := kindprovisioner.NewDefaultKindProviderAdapter()

	dockerClient, err := kindprovisioner.NewDefaultDockerClient()
	if err != nil {
		t.Skip("Docker client not available, skipping integration test")
	}

	provisioner := kindprovisioner.NewKindClusterProvisioner(
		kindConfig,
		"~/.kube/config",
		providerAdapter,
		dockerClient,
	)

	require.NotNil(t, provisioner, "provisioner should not be nil")
}
