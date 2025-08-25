package dockerprovisioner_test

import (
	"context"
	"errors"
	"testing"

	dockerprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/container_engine/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var errBoom = errors.New("boom")

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
	mockClient := dockerprovisioner.NewMockAPIClient(t)

	// Act
	provisioner := dockerprovisioner.NewDockerProvisioner(mockClient)

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
	assert.Contains(t, err.Error(), "docker ping failed")
	assert.Contains(t, err.Error(), "boom")
	mockClient.AssertExpectations(t)
}

// newProvisionerForTest creates a DockerProvisioner with mocked dependencies for testing.
func newProvisionerForTest(t *testing.T) (*dockerprovisioner.DockerProvisioner, *dockerprovisioner.MockAPIClient) {
	t.Helper()
	mockClient := dockerprovisioner.NewMockAPIClient(t)
	provisioner := dockerprovisioner.NewDockerProvisioner(mockClient)

	return provisioner, mockClient
}
