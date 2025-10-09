// Package k9s provides a k9s client implementation.
package k9s

import (
	"os"

	k9scmd "github.com/derailed/k9s/cmd"
	"github.com/spf13/cobra"
)

// Executor defines the interface for executing k9s.
type Executor interface {
	Execute()
}

// DefaultK9sExecutor is the default implementation that calls k9s directly.
type DefaultK9sExecutor struct{}

// Execute calls the k9s cmd.Execute function.
func (e *DefaultK9sExecutor) Execute() {
	k9scmd.Execute()
}

// Client wraps k9s command functionality.
type Client struct {
	executor Executor
}

// NewClient creates a new k9s client instance with the default executor.
func NewClient() *Client {
	return &Client{
		executor: &DefaultK9sExecutor{},
	}
}

// NewClientWithExecutor creates a new k9s client with a custom executor for testing.
func NewClientWithExecutor(executor Executor) *Client {
	return &Client{
		executor: executor,
	}
}

// CreateConnectCommand creates a k9s command with all its flags and behavior.
func (c *Client) CreateConnectCommand(kubeConfigPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to cluster with k9s",
		Long:  "Launch k9s terminal UI to interactively manage your Kubernetes cluster.",
		RunE: func(_ *cobra.Command, args []string) error {
			return c.runK9s(kubeConfigPath, args)
		},
		SilenceUsage: true,
	}

	return cmd
}

func (c *Client) runK9s(kubeConfigPath string, args []string) error {
	// Set up os.Args to pass flags to k9s
	originalArgs := os.Args

	defer func() {
		os.Args = originalArgs
	}()

	// Build arguments for k9s
	k9sArgs := []string{"k9s"}

	// Add kubeconfig flag if provided
	if kubeConfigPath != "" {
		k9sArgs = append(k9sArgs, "--kubeconfig", kubeConfigPath)
	}

	// Append all additional arguments passed by user
	k9sArgs = append(k9sArgs, args...)

	// Set os.Args for k9s to parse
	os.Args = k9sArgs

	// Execute k9s using the executor (allows mocking for tests)
	c.executor.Execute()

	return nil
}
