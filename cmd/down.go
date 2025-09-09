// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return config.NewCobraCommand(
		"down",
		"Stop and remove the Kubernetes cluster",
		`Stop and remove the Kubernetes cluster defined in the project configuration.`,
		handleDownRunE,
		[]config.FieldSelector{}, // No specific configuration flags needed
	)
}

// handleDownRunE handles the down command.
func handleDownRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster stopped and removed successfully (stub implementation)")

	return nil
}
