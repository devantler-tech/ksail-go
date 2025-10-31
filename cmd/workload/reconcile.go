package workload

import (
	"github.com/devantler-tech/ksail-go/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates the workload reconcile command.
func NewReconcileCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reconcile",
		Short:        "Reconcile workloads with the cluster",
		Long:         "Trigger reconciliation tooling to sync local workloads with your cluster.",
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(runtimeContainer)

	return cmd
}
