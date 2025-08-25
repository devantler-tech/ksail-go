package podmanprovisioner_test

import (
	"context"
	"errors"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/provisioner"
	podmanprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/podman"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

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

	expectedPing := types.Ping{
		APIVersion:     "1.41",
		OSType:         "",
		Experimental:   false,
		SwarmStatus:    nil,
		BuilderVersion: "",
	}
	mockClient.On("Ping", mock.MatchedBy(func(_ context.Context) bool {
		return true
	})).Return(expectedPing, nil)

	// Act
	ready, err := provisioner.CheckReady()

	// Assert
	require.NoError(t, err)
	assert.True(t, ready)
	mockClient.AssertExpectations(t)
}

func TestCheckReady_Error_PingFailed(t *testing.T) {
	t.Parallel()

	// Arrange
	provisioner, mockClient := newProvisionerForTest(t)

	mockClient.On("Ping", mock.MatchedBy(func(_ context.Context) bool {
		return true
	})).Return(types.Ping{
		APIVersion:     "",
		OSType:         "",
		Experimental:   false,
		BuilderVersion: "",
		SwarmStatus:    nil,
	}, errBoom)

	// Act
	ready, err := provisioner.CheckReady()

	// Assert
	require.Error(t, err)
	assert.False(t, ready)
	assert.Contains(t, err.Error(), "podman ping failed")
	assert.Contains(t, err.Error(), "boom")
	mockClient.AssertExpectations(t)
}

// newProvisionerForTest creates a PodmanProvisioner with mocked dependencies for testing.
func newProvisionerForTest(t *testing.T) (*podmanprovisioner.PodmanProvisioner, *provisioner.MockAPIClient) {
	t.Helper()
	mockClient := provisioner.NewMockAPIClient(t)
	provisioner := podmanprovisioner.NewPodmanProvisioner(mockClient)

	return provisioner, mockClient
}
