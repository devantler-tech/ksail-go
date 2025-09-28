// Package workload provides the workload command namespace.
package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

const (
	placeholderReconcileMessage = "Workload reconciliation coming soon."
	placeholderApplyMessage     = "Workload apply coming soon."
	placeholderInstallMessage   = "Workload install coming soon."
)

// NewWorkloadCmd creates and returns the workload command group namespace.
func NewWorkloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workload",
		Short: "Manage workload operations",
		Long: "Group workload commands under a single namespace to reconcile, " +
			"apply, or install workloads.",
		SilenceErrors:              true,
		SilenceUsage:               true,
		SuggestionsMinimumDistance: cmdhelpers.SuggestionsMinimumDistance,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newWorkloadReconcileCmd())
	cmd.AddCommand(newWorkloadApplyCmd())
	cmd.AddCommand(newWorkloadInstallCmd())

	return cmd
}

func newWorkloadReconcileCmd() *cobra.Command {
	return newPlaceholderCommand(
		"reconcile",
		"Reconcile workloads with the cluster",
		"Trigger reconciliation tooling to sync local workloads with your cluster.",
		placeholderReconcileMessage,
	)
}

func newWorkloadApplyCmd() *cobra.Command {
	return newPlaceholderCommand(
		"apply",
		"Apply manifests",
		"Apply local Kubernetes manifests to your cluster.",
		placeholderApplyMessage,
	)
}

func newWorkloadInstallCmd() *cobra.Command {
	return newPlaceholderCommand(
		"install",
		"Install Helm charts",
		"Install Helm charts to provision workloads through KSail.",
		placeholderInstallMessage,
	)
}

func newPlaceholderCommand(use, short, long, message string) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        use,
		Short:                      short,
		Long:                       long,
		SilenceErrors:              true,
		SilenceUsage:               true,
		SuggestionsMinimumDistance: cmdhelpers.SuggestionsMinimumDistance,
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.Infoln(cmd.OutOrStdout(), message)

			return nil
		},
	}

	return cmd
}
