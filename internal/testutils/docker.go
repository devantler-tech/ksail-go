package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/shared"
	"github.com/docker/docker/client"
)

// StubDockerClientFailure stubs the Docker client factory in internal/shared to return an error.
// This is useful for testing error handling in code that uses shared.WithDockerClient.
//
// The function automatically restores the original factory using t.Cleanup.
//
// Example usage:
//
//	testutils.StubDockerClientFailure(t, errDockerClientFailure)
//	err := shared.WithDockerClient(cmd, func(client.APIClient) error { return nil })
//	require.Error(t, err)
func StubDockerClientFailure(t *testing.T, err error) {
	t.Helper()

	// Get the current factory to restore later
	originalFactory := shared.GetDockerClientFactory()

	// Set the factory to return the test error
	shared.SetDockerClientFactory(func(...client.Opt) (*client.Client, error) {
		return nil, err
	})

	// Restore original factory on cleanup
	t.Cleanup(func() {
		shared.SetDockerClientFactory(originalFactory)
	})
}
