package cmd

import (
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		func(cmd *cobra.Command, manager *configmanager.ConfigManager, args []string) error {
			return cmdhelpers.StandardClusterCommandRunE(
				"cluster destroyed successfully",
			)(cmd, manager, args)
		},
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}
