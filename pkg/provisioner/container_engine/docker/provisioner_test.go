package dockerprovisioner_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	containerengineprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine"
	dockerprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/docker"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/testutils"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDockerProvisioner_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	cli := CreateDockerClient(t)

	// Act
	provisioner := dockerprovisioner.NewDockerProvisioner(cli)

	// Assert
	assert.NotNil(t, provisioner)
}

func CreateDockerClient(t *testing.T) *client.Client {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	return cli
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
	testutils.TestCheckReadySuccess(
		t,
		func(
			mockClient *provisioner.MockAPIClient,
		) containerengineprovisioner.ContainerEngineProvisioner {
			return dockerprovisioner.NewDockerProvisioner(mockClient)
		},
	)
}

func TestCheckReady_Error_PingFailed(t *testing.T) {
	t.Parallel()
	testutils.TestCheckReadyError(
		t,
		func(
			mockClient *provisioner.MockAPIClient,
		) containerengineprovisioner.ContainerEngineProvisioner {
			return dockerprovisioner.NewDockerProvisioner(mockClient)
		},
		"docker ping failed",
	)
}
