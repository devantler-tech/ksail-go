package podmanprovisioner_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	containerengineprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine"
	podmanprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/podman"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPodmanProvisioner_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	// Act
	provisioner := podmanprovisioner.NewPodmanProvisioner(cli)

	// Assert
	assert.NotNil(t, provisioner)
}

func TestNewPodmanProvisioner_WithMockClient(t *testing.T) {
	t.Parallel()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)

	// Act
	provisioner := podmanprovisioner.NewPodmanProvisioner(mockClient)

	// Assert
	assert.NotNil(t, provisioner)
}

func TestCheckReady_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner, mockClient := newProvisionerForTest(t)

	// Act & Assert
	containerengineprovisioner.TestCheckReadySuccess(t, provisioner, mockClient)
}

func TestCheckReady_Error_PingFailed(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner, mockClient := newProvisionerForTest(t)

	// Act & Assert
	containerengineprovisioner.TestCheckReadyError(t, provisioner, mockClient, "podman ping failed")
}

// newProvisionerForTest creates a PodmanProvisioner with mocked dependencies for testing.
func newProvisionerForTest(t *testing.T) (*podmanprovisioner.PodmanProvisioner, *provisioner.MockAPIClient) {
	t.Helper()
	mockClient := provisioner.NewMockAPIClient(t)
	provisioner := podmanprovisioner.NewPodmanProvisioner(mockClient)

	return provisioner, mockClient
}
