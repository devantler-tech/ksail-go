// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return config.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		handleStopRunE,
		[]config.FieldSelector{}, // No specific configuration flags needed
	)
}

// handleStopRunE handles the stop command.
func handleStopRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster stopped successfully (stub implementation)")

	return nil
}
