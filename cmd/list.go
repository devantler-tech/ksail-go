// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/inputs"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kubernetes clusters",
		Long:  `List all Kubernetes clusters managed by KSail.`,
		RunE:  handleListRunE,
	}

	// Add flags
	inputs.AddListFlags(cmd)

	return cmd
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