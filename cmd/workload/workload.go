// Package workload provides the workload command namespace.
package workload

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewWorkloadCmd creates and returns the workload command group namespace.
func NewWorkloadCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
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

	cmd.AddCommand(NewReconcileCmd(runtimeContainer))
	cmd.AddCommand(NewApplyCmd(runtimeContainer))
	cmd.AddCommand(NewInstallCmd(runtimeContainer))
	cmd.AddCommand(NewGetCmd(runtimeContainer))

	return cmd
}
