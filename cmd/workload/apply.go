package workload

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewApplyCmd creates the workload apply command.
func NewApplyCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"apply",
		"Apply manifests",
		"Apply local Kubernetes manifests to your cluster.",
		HandleApplyRunE,
	)
}

// HandleApplyRunE handles the apply command.
func HandleApplyRunE(
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
