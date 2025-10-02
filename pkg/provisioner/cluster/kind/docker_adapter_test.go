package kindprovisioner_test

import (
	"testing"

	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestNewDefaultDockerClient(t *testing.T) {
	t.Parallel()

	dockerClient, err := kindprovisioner.NewDefaultDockerClient()
	// Docker may not be available in all test environments
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	require.NotNil(t, dockerClient, "docker client should not be nil")
	require.NoError(t, err, "should not error when Docker is available")
}

func TestDefaultDockerClientImplementsInterface(t *testing.T) {
	t.Parallel()

	dockerClient, err := kindprovisioner.NewDefaultDockerClient()
	// Docker may not be available in all test environments
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}

	// Verify that the returned client implements ContainerAPIClient interface
	require.NotNil(t, dockerClient)
}

func TestDefaultDockerClientUsageInProvisioner(t *testing.T) {
	t.Parallel()

	// Test that the Docker client can be used with the provisioner
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

	assert.NotNil(t, provisioner, "provisioner should not be nil")
}
