package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates the workload reconcile command.
func NewReconcileCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"reconcile",
		"Reconcile workloads with the cluster",
		"Trigger reconciliation tooling to sync local workloads with your cluster.",
		helpers.HandleConfigLoadRunE,
	)
}
