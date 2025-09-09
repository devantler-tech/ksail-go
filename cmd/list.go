// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	cmd := config.NewCobraCommand(
		"list",
		"List Kubernetes clusters",
		`List all Kubernetes clusters managed by KSail.`,
		handleListRunE,
	)
	
	// Add the special --all flag manually since it's CLI-only
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
	
	return cmd
}

// handleListRunE handles the list command.
func handleListRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Bind the --all flag manually since it's added after command creation
	_ = configManager.GetViper().BindPFlag("all", cmd.Flags().Lookup("all"))
	
	all := configManager.GetViper().GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters (stub implementation)")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters (stub implementation)")
	}

	return nil
}
