// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	return config.NewCobraCommand(
		"up",
		"Start the Kubernetes cluster",
		`Start the Kubernetes cluster defined in the project configuration.`,
		handleUpRunE,
	)
}

// handleUpRunE handles the up command.
func handleUpRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(
		cmd.OutOrStdout(),
		"Cluster created and started successfully (stub implementation)",
	)

	return nil
}
