// Package k9s provides a k9s client implementation.
package k9s

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

// Client wraps k9s command functionality.
type Client struct {
	k9sBinaryPath string
}

// NewClient creates a new k9s client instance.
func NewClient(k9sBinaryPath string) *Client {
	if k9sBinaryPath == "" {
		k9sBinaryPath = "k9s"
	}

	return &Client{
		k9sBinaryPath: k9sBinaryPath,
	}
}

// CreateConnectCommand creates a k9s command with all its flags and behavior.
func (c *Client) CreateConnectCommand(kubeConfigPath string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to cluster with k9s",
		Long:  "Launch k9s terminal UI to interactively manage your Kubernetes cluster.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runK9s(cmd.Context(), kubeConfigPath, args)
		},
		DisableFlagParsing: true,
		SilenceUsage:       true,
	}

	return cmd
}

func (c *Client) runK9s(ctx context.Context, kubeConfigPath string, args []string) error {
	k9sArgs := []string{}

	// Add kubeconfig flag if provided
	if kubeConfigPath != "" {
		k9sArgs = append(k9sArgs, "--kubeconfig", kubeConfigPath)
	}

	// Append all additional arguments passed by user
	k9sArgs = append(k9sArgs, args...)

	// #nosec G204 -- k9sBinaryPath is controlled by NewClient, not user input
	k9sCmd := exec.CommandContext(ctx, c.k9sBinaryPath, k9sArgs...)
	k9sCmd.Stdin = os.Stdin
	k9sCmd.Stdout = os.Stdout
	k9sCmd.Stderr = os.Stderr

	err := k9sCmd.Run()
	if err != nil {
		return fmt.Errorf("run k9s: %w", err)
	}

	return nil
}
