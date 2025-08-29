package podmanprovisioner_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	containerengineprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine"
	podmanprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/podman"
	"github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/testutils"
	"github.com/stretchr/testify/assert"
)

func TestNewPodmanProvisioner_Success(t *testing.T) {
	t.Parallel()

	// Arrange
	cli := testutils.CreateTestDockerClient(t)

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
	testutils.TestCheckReadySuccess(
		t,
		func(
			mockClient *provisioner.MockAPIClient,
		) containerengineprovisioner.ContainerEngineProvisioner {
			return podmanprovisioner.NewPodmanProvisioner(mockClient)
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
			return podmanprovisioner.NewPodmanProvisioner(mockClient)
		},
		"podman ping failed",
	)
}
