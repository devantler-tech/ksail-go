// Package workload provides the workload command namespace.
package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/spf13/cobra"
)

// NewWorkloadCmd creates and returns the workload command group namespace.
func NewWorkloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workload",
		Short: "Manage workload operations",
		Long: "Group workload commands under a single namespace to reconcile, " +
			"apply, or install workloads.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.SuggestionsMinimumDistance = helpers.SuggestionsMinimumDistance

	cmd.AddCommand(NewReconcileCmd())
	cmd.AddCommand(NewApplyCmd())
	cmd.AddCommand(NewInstallCmd())

	return cmd
}
