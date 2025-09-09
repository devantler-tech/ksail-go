// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	return factory.NewCobraCommandWithSelectiveBinding(
		"list",
		"List Kubernetes clusters",
		`List all Kubernetes clusters managed by KSail.`,
		handleListRunE,
		[]string{"all"}, // Only include the 'all' flag for list command
	)
}

// handleListRunE handles the list command.
func handleListRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	all := configManager.GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters (stub implementation)")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters (stub implementation)")
	}

	return nil
}
