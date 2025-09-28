package workload

import (
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

const reconcileMessage = "Workload reconciliation coming soon."

// NewReconcileCommand creates the workload reconcile command.
func NewReconcileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile workloads with the cluster",
		Long:  "Trigger reconciliation tooling to sync local workloads with your cluster.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			notify.Infoln(cmd.OutOrStdout(), reconcileMessage)

			return nil
		},
	}

	applyCommonCommandConfig(cmd)

	return cmd
}
