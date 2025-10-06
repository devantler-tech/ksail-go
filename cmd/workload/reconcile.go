package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/runtime"
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates the workload reconcile command.
func NewReconcileCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reconcile",
		Short:        "Reconcile workloads with the cluster",
		Long:         "Trigger reconciliation tooling to sync local workloads with your cluster.",
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(rt)

	return cmd
}
