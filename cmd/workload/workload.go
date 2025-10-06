// Package workload provides the workload command namespace.
package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/runtime"
	"github.com/spf13/cobra"
)

// NewWorkloadCmd creates and returns the workload command group namespace.
func NewWorkloadCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workload",
		Short: "Manage workload operations",
		Long: "Group workload commands under a single namespace to reconcile, " +
			"apply, or install workloads.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
		SilenceUsage: true,
	}

	cmd.AddCommand(NewReconcileCmd(rt))
	cmd.AddCommand(NewApplyCmd(rt))
	cmd.AddCommand(NewInstallCmd(rt))

	return cmd
}
