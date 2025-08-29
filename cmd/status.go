// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of the Kubernetes cluster",
		Long:  `Show the current status of the Kubernetes cluster.`,
		RunE:  handleStatusRunE,
	}

	return cmd
}

// handleStatusRunE handles the status command.
func handleStatusRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster status: Running (stub implementation)")

	return nil
}
