// Package k9s provides a k9s client implementation.
//
// Coverage Note: The DefaultK9sExecutor.Execute() method and parts of the
// HandleConnectRunE execution path cannot be fully tested in unit tests because they
// require launching k9s which needs an actual terminal UI. These paths are validated
// through:
// - Integration testing with actual k9s installation
// - Manual verification of the connect command
// - Mock-based testing of all logic leading up to k9s execution
package k9s

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// Executor defines the interface for executing k9s.
type Executor interface {
	Execute(args []string) error
}

// DefaultK9sExecutor is the default implementation that calls k9s as an external command.
type DefaultK9sExecutor struct{}

// Execute runs k9s as an external command with the provided arguments.
func (e *DefaultK9sExecutor) Execute(args []string) error {
	// Check if k9s is installed
	k9sPath, err := exec.LookPath("k9s")
	if err != nil {
		return fmt.Errorf("k9s not found in PATH - please install k9s: https://k9scli.io/topics/install/")
	}

	// Create command
	cmd := exec.Command(k9sPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run k9s
	return cmd.Run()
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
func (c *Client) CreateConnectCommand(kubeConfigPath, context string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to cluster with k9s",
		Long:  "Launch k9s terminal UI to interactively manage your Kubernetes cluster.",
		RunE: func(_ *cobra.Command, args []string) error {
			return c.runK9s(kubeConfigPath, context, args)
		},
		SilenceUsage: true,
	}

	return cmd
}

func (c *Client) runK9s(kubeConfigPath, context string, args []string) error {
	// Build arguments for k9s
	k9sArgs := []string{}

	// Add kubeconfig flag if provided
	if kubeConfigPath != "" {
		k9sArgs = append(k9sArgs, "--kubeconfig", kubeConfigPath)
	}

	// Add context flag if provided
	if context != "" {
		k9sArgs = append(k9sArgs, "--context", context)
	}

	// Append all additional arguments passed by user
	k9sArgs = append(k9sArgs, args...)

	// Execute k9s using the executor (allows mocking for tests)
	return c.executor.Execute(k9sArgs)
}
