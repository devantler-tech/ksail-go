package testutils

import (
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

// CreateTestDockerClient creates a new Docker client for use in tests.
func CreateTestDockerClient(t *testing.T) *client.Client {
	t.Helper()

	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)

	return cli
}
