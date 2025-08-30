// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	return factory.NewCobraCommandWithFlags(
		"list",
		"List Kubernetes clusters",
		`List all Kubernetes clusters managed by KSail.`,
		handleListRunE,
		func(cmd *cobra.Command) {
			cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
		},
	)
}

// handleListRunE handles the list command.
func handleListRunE(cmd *cobra.Command, _ []string) error {
	all, _ := cmd.Flags().GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters (stub implementation)")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters (stub implementation)")
	}

	return nil
}
