package shared

import (
	"testing"

	"github.com/docker/docker/client"
)

// StubDockerClientFailure overrides the default docker client factory to return an error.
// This is exported for use in other test packages that need to test docker client failures.
func StubDockerClientFailure(t *testing.T, err error) {
	t.Helper()

	original := dockerClientFactory

	t.Cleanup(func() {
		dockerClientFactory = original
	})

	dockerClientFactory = func(...client.Opt) (*client.Client, error) {
		return nil, err
	}
}
