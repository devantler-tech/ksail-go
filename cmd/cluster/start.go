package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		HandleStartRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleStartRunE handles the start command.
func HandleStartRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	err := cmdhelpers.ExecuteTimedClusterCommand(
		cmd,
		manager,
		"Cluster started successfully (stub implementation)",
	)
	if err != nil {
		return fmt.Errorf("failed to execute start command: %w", err)
	}

	return nil
}
