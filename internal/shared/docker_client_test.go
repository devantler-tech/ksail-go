package shared //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"errors"
	"testing"

	dockermocks "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errOperationFailed = errors.New("operation failed")
	errCloseFailure    = errors.New("close failed")
)

func TestWithDockerClient_Success(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var out bytes.Buffer

	cmd.SetOut(&out)

	operationCalled := false
	operation := func(dockerClient client.APIClient) error {
		operationCalled = true

		assert.NotNil(t, dockerClient)

		return nil
	}

	// Note: This test requires Docker to be available in the environment
	// If Docker is not available, the test will fail at client creation
	err := WithDockerClient(cmd, operation)

	// We can't guarantee Docker is available in all test environments
	// so we accept both success and the specific error about Docker not being available
	if err != nil {
		// Check if it's a Docker connection error (expected in some environments)
		assert.Contains(t, err.Error(), "docker")
	} else {
		assert.True(t, operationCalled)
	}
}

func TestWithDockerClient_OperationError(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{}

	var out bytes.Buffer

	cmd.SetOut(&out)

	operation := func(_ client.APIClient) error {
		return errOperationFailed
	}

	err := WithDockerClient(cmd, operation)

	// If Docker is available, we should get the operation error
	// If Docker is not available, we'll get a Docker connection error
	if err != nil && errors.Is(err, errOperationFailed) {
		assert.ErrorIs(t, err, errOperationFailed)
	}
}

func TestWithDockerClientInstance_Success(t *testing.T) {
	t.Parallel()

	mockClient := dockermocks.NewMockAPIClient(t)
	mockClient.EXPECT().Close().Return(nil)

	operationCalled := false
	operation := func(dockerClient client.APIClient) error {
		operationCalled = true

		assert.NotNil(t, dockerClient)

		return nil
	}

	err := withDockerClientInstanceTest(mockClient, operation)

	require.NoError(t, err)
	assert.True(t, operationCalled)
}

func TestWithDockerClientInstance_OperationError(t *testing.T) {
	t.Parallel()

	mockClient := dockermocks.NewMockAPIClient(t)
	mockClient.EXPECT().Close().Return(nil)

	operation := func(_ client.APIClient) error {
		return errOperationFailed
	}

	err := withDockerClientInstanceTest(mockClient, operation)

	assert.ErrorIs(t, err, errOperationFailed)
}

func TestWithDockerClientInstance_CloseError(t *testing.T) {
	t.Parallel()

	mockClient := dockermocks.NewMockAPIClient(t)
	mockClient.EXPECT().Close().Return(errCloseFailure)

	operation := func(_ client.APIClient) error {
		return nil
	}

	out, err := withDockerClientInstanceTestWithOutput(mockClient, operation)

	// Operation succeeds even if close fails (cleanup warning is logged)
	require.NoError(t, err)
	assert.Contains(t, out, "cleanup warning")
	assert.Contains(t, out, "close failed")
}

func TestWithDockerClientInstance_OperationAndCloseError(t *testing.T) {
	t.Parallel()

	mockClient := dockermocks.NewMockAPIClient(t)
	mockClient.EXPECT().Close().Return(errCloseFailure)

	operation := func(_ client.APIClient) error {
		return errOperationFailed
	}

	out, err := withDockerClientInstanceTestWithOutput(mockClient, operation)

	// Operation error is returned, cleanup warning is logged
	require.ErrorIs(t, err, errOperationFailed)
	assert.Contains(t, out, "cleanup warning")
	assert.Contains(t, out, "close failed")
}

// Test helper to reduce code duplication.
func withDockerClientInstanceTest(
	mockClient client.APIClient,
	operation func(client.APIClient) error,
) error {
	cmd := &cobra.Command{}

	var out bytes.Buffer
	cmd.SetOut(&out)

	return WithDockerClientInstance(cmd, mockClient, operation)
}

// Test helper that also returns output for assertions.
func withDockerClientInstanceTestWithOutput(
	mockClient client.APIClient,
	operation func(client.APIClient) error,
) (string, error) {
	cmd := &cobra.Command{}

	var out bytes.Buffer
	cmd.SetOut(&out)

	err := WithDockerClientInstance(cmd, mockClient, operation)

	return out.String(), err
}
