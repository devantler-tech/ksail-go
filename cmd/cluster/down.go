package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		HandleDownRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleDownRunE handles the down command.
func HandleDownRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	err := cmdhelpers.ExecuteTimedClusterCommand(
		cmd,
		manager,
		"Cluster stopped and deleted successfully (stub implementation)",
	)
	if err != nil {
		return fmt.Errorf("failed to provision cluster down: %w", err)
	}

	return nil
}
