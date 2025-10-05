package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates the workload reconcile command.
func NewReconcileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile workloads with the cluster",
		Long:  "Trigger reconciliation tooling to sync local workloads with your cluster.",
		RunE:  utils.HandleConfigLoadRunE,
	}
}
