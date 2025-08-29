// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the Kubernetes cluster",
		Long:  `Stop the Kubernetes cluster without removing it.`,
		RunE:  handleStopRunE,
	}

	return cmd
}

// handleStopRunE handles the stop command.
func handleStopRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster stopped successfully (stub implementation)")

	return nil
}
