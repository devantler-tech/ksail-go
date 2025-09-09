// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return config.NewCobraCommand(
		"status",
		"Show status of the Kubernetes cluster",
		`Show the current status of the Kubernetes cluster.`,
		handleStatusRunE,
		[]string{}, // No specific configuration flags needed
	)
}

// handleStatusRunE handles the status command.
func handleStatusRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster status: Running (stub implementation)")

	return nil
}
