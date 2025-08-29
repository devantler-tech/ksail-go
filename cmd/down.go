// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Stop and remove the Kubernetes cluster",
		Long:  `Stop and remove the Kubernetes cluster defined in the project configuration.`,
		RunE:  handleDownRunE,
	}

	return cmd
}

// handleDownRunE handles the down command.
func handleDownRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster stopped and removed successfully (stub implementation)")

	return nil
}
