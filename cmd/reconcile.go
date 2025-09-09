// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates and returns the reconcile command.
func NewReconcileCmd() *cobra.Command {
	return config.NewCobraCommand(
		"reconcile",
		"Reconcile workloads in the cluster",
		`Reconcile workloads in the cluster to match the desired state
defined in configuration files.`,
		handleReconcileRunE,
	)
}

// handleReconcileRunE handles the reconcile command.
func handleReconcileRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Workloads reconciled successfully (stub implementation)")

	return nil
}
