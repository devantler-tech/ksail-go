// Package workload provides the workload command namespace.
package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
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

	applyCommonCommandConfig(cmd)

	cmd.AddCommand(NewReconcileCommand())
	cmd.AddCommand(NewApplyCommand())
	cmd.AddCommand(NewInstallCommand())

	return cmd
}

func applyCommonCommandConfig(cmd *cobra.Command) {
	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	cmd.SuggestionsMinimumDistance = cmdhelpers.SuggestionsMinimumDistance
}
