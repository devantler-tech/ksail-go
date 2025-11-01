package shared

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

type dockerClientFactoryFunc func(opts ...client.Opt) (*client.Client, error)

func defaultDockerClientFactory(opts ...client.Opt) (*client.Client, error) {
	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("creating docker client: %w", err)
	}

	return cli, nil
}

//nolint:gochecknoglobals // Allow tests to override Docker client creation.
var dockerClientFactory dockerClientFactoryFunc = defaultDockerClientFactory

// SetDockerClientFactoryForTest allows tests to override the Docker client factory.
// This should only be used in tests.
func SetDockerClientFactoryForTest(
	factory func(opts ...client.Opt) (*client.Client, error),
) func() {
	original := dockerClientFactory
	dockerClientFactory = factory

	return func() {
		dockerClientFactory = original
	}
}

// WithDockerClient creates a Docker client, executes the given function, and cleans up.
// Returns an error if client creation fails or if the function returns an error.
func WithDockerClient(cmd *cobra.Command, operation func(client.APIClient) error) error {
	dockerClient, err := dockerClientFactory(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

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
