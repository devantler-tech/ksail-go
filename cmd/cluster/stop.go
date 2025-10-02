package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		HandleStopRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleStopRunE handles the stop command.
func HandleStopRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	return cmdhelpers.ExecuteTimedClusterCommand(
		cmd,
		manager,
		"Cluster stopped successfully (stub implementation)",
		false,
	)
}
