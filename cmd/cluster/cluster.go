package cluster

import (
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/spf13/cobra"
)

// NewClusterCmd creates the parent cluster command and wires lifecycle subcommands beneath it.
func NewClusterCmd() *cobra.Command {
	cmd := cmdhelpers.NewCobraCommand(
		"cluster",
		"Manage cluster lifecycle commands",
		`Manage lifecycle operations for local Kubernetes clusters, including provisioning, teardown, and status.`,
		handleClusterRunE,
	)

	cmd.AddCommand(NewUpCmd())
	cmd.AddCommand(NewDownCmd())
	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewStopCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewListCmd())

	return cmd
}

func handleClusterRunE(cmd *cobra.Command, _ *configmanager.ConfigManager, _ []string) error {
	// Cobra help cannot fail at runtime, so ignoring the error is safe.
	_ = cmd.Help()

	return nil
}
