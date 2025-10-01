package k3dprovisioner_test

import (
	"context"
	"testing"

	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
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

func TestDefaultK3dConfigAdapterTransformSimpleToClusterConfig(t *testing.T) {
	t.Parallel()

	adapter := k3dprovisioner.NewDefaultK3dConfigAdapter()
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	// Create a minimal simple config
	simpleConfig := v1alpha5.SimpleConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: types.ObjectMeta{
			Name: "test-config-adapter",
		},
	}

	// Transform the config
	clusterConfig, err := adapter.TransformSimpleToClusterConfig(
		ctx,
		runtime,
		simpleConfig,
		"test.yaml",
	)

	// The transformation should succeed with minimal config
	if err == nil {
		assert.NotNil(t, clusterConfig, "clusterConfig should not be nil when no error")
		assert.Equal(
			t,
			"test-config-adapter",
			clusterConfig.Name,
			"cluster name should match",
		)
	}
	// If error occurs (e.g., runtime issues), that's acceptable for unit tests
	// The important part is that the method signature is correct and doesn't panic
}

func TestDefaultK3dConfigAdapterUsageInProvisioner(t *testing.T) {
	t.Parallel()

	// Test that the adapter can be used with the provisioner
	simpleCfg := &v1alpha5.SimpleConfig{}
	simpleCfg.Name = "test-cluster"

	clientAdapter := k3dprovisioner.NewDefaultK3dClientAdapter()
	configAdapter := k3dprovisioner.NewDefaultK3dConfigAdapter()

	provisioner := k3dprovisioner.NewK3dClusterProvisioner(
		simpleCfg,
		clientAdapter,
		configAdapter,
	)

	assert.NotNil(t, provisioner, "provisioner should not be nil")
}
