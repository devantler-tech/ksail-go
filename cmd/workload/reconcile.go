package workload

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates the workload reconcile command.
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
	_ *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	tmr := timer.New()
	tmr.Start()

	_, err := manager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return nil
}
