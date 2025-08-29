// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start the Kubernetes cluster",
		Long:  `Start the Kubernetes cluster defined in the project configuration.`,
		RunE:  handleUpRunE,
	}

	return cmd
}

// handleUpRunE handles the up command.
func handleUpRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster created and started successfully (stub implementation)")

	return nil
}
