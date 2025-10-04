package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewApplyCommand creates the workload apply command.
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
	cmd *cobra.Command,
	_ *configmanager.ConfigManager,
	_ []string,
) error {
	return nil
}
