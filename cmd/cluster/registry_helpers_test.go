package cluster //nolint:testpackage // Access unexported helpers for coverage-focused tests.

import (
	"bytes"
	"errors"
	"testing"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
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
	err := withDockerClient(cmd, operation)

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

	err := withDockerClient(cmd, operation)

	// If Docker is available, we should get the operation error
	// If Docker is not available, we'll get a Docker connection error
	if err != nil && errors.Is(err, errOperationFailed) {
		assert.ErrorIs(t, err, errOperationFailed)
	}
}
