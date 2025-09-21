// Package cmd provides the command-line interface for KSail.
package cmd

import (
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		func(cmd *cobra.Command, manager *configmanager.ConfigManager, args []string) error {
			return cmdhelpers.StandardClusterCommandRunE(
				"Cluster started successfully (stub implementation)",
			)(cmd, manager, args)
		},
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}
