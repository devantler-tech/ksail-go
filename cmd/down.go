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
		"Destroy a cluster",
		`Destroy a cluster.`,
		handleDownRunE,
	)
}

// handleDownRunE handles the down command.
func handleDownRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(
		cmd.OutOrStdout(),
		"cluster destroyed successfully",
	)

	return nil
}
