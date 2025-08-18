// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewUpdateCmd creates and returns the update command.
func NewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the Kubernetes cluster",
		Long:  `Update the Kubernetes cluster configuration and workloads.`,
		RunE:  handleUpdateRunE,
	}

	return cmd
}

// handleUpdateRunE handles the update command.
func handleUpdateRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster updated successfully (stub implementation)")

	return nil
}