package dockerprovisioner_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	dockerprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/docker"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/testutils"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDockerProvisioner_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	// Act
	provisioner := dockerprovisioner.NewDockerProvisioner(cli)

	// Assert
	assert.NotNil(t, provisioner)
}

func TestNewDockerProvisioner_WithMockClient(t *testing.T) {
	t.Parallel()

	// Arrange
	mockClient := provisioner.NewMockAPIClient(t)

	// Act
	provisioner := dockerprovisioner.NewDockerProvisioner(mockClient)

	// Assert
	assert.NotNil(t, provisioner)
}

func TestCheckReady_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner, mockClient := newProvisionerForTest(t)

	// Act & Assert
	testutils.TestCheckReadySuccess(t, provisioner, mockClient)
}

func TestCheckReady_Error_PingFailed(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner, mockClient := newProvisionerForTest(t)

	// Act & Assert
	testutils.TestCheckReadyError(t, provisioner, mockClient, "docker ping failed")
}

// newProvisionerForTest creates a DockerProvisioner with mocked dependencies for testing.
func newProvisionerForTest(t *testing.T) (*dockerprovisioner.DockerProvisioner, *provisioner.MockAPIClient) {
	t.Helper()
	mockClient := provisioner.NewMockAPIClient(t)
	provisioner := dockerprovisioner.NewDockerProvisioner(mockClient)

	return provisioner, mockClient
}
