package shared

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

// WithDockerClient creates a Docker client, executes the given function, and cleans up.
// Returns an error if client creation fails or if the function returns an error.
// For testing, use WithDockerClientInstance with a mock client instead.
func WithDockerClient(cmd *cobra.Command, operation func(client.APIClient) error) error {
	dockerClient, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	return WithDockerClientInstance(cmd, dockerClient, operation)
}

// WithDockerClientInstance executes the given function with a provided Docker client and handles cleanup.
// This is useful for testing with mock clients.
// The client will be closed after the operation completes.
func WithDockerClientInstance(
	cmd *cobra.Command,
	dockerClient client.APIClient,
	operation func(client.APIClient) error,
) error {
	defer func() {
		closeErr := dockerClient.Close()
		if closeErr != nil {
			// Log cleanup error but don't fail the operation
			notify.WriteMessage(notify.Message{
				Type: notify.WarningType,
				Content: fmt.Sprintf(
					"cleanup warning: failed to close docker client: %v",
					closeErr,
				),
				Writer: cmd.OutOrStdout(),
			})
		}
	}()

	return operation(dockerClient)
}
