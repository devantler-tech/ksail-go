package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewReconcileCommand creates the workload reconcile command.
func NewReconcileCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"reconcile",
		"Reconcile workloads with the cluster",
		"Trigger reconciliation tooling to sync local workloads with your cluster.",
		HandleReconcileRunE,
	)
}

// HandleReconcileRunE handles the reconcile command.
func HandleReconcileRunE(
	cmd *cobra.Command,
	_ *configmanager.ConfigManager,
	_ []string,
) error {
	cmd.Println("ℹ Workload reconciliation coming soon.")

	return nil
}
