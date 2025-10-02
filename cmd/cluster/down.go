package cluster

import (
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
	return cmdhelpers.ExecuteTimedClusterCommand(
		cmd,
		manager,
		"cluster destroyed successfully",
		false,
	)
}
