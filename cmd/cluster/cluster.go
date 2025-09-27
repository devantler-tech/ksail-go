package cluster

import (
	"fmt"

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
	// Cobra Help() can return an error (e.g., output stream or template issues); wrap it for clarity.
	err := cmd.Help()
	if err != nil {
		return fmt.Errorf("displaying cluster command help: %w", err)
	}

	return nil
}
