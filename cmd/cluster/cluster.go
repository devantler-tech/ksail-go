package cluster

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewClusterCmd creates the parent cluster command and wires lifecycle subcommands beneath it.
func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cluster",
		Short:        "Manage cluster lifecycle",
		Long:         `Manage lifecycle operations for local Kubernetes clusters, including provisioning, teardown, and status.`,
		RunE:         handleClusterRunE,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewCreateCmd())
	cmd.AddCommand(NewDeleteCmd())
	cmd.AddCommand(NewStartCmd())
	cmd.AddCommand(NewStopCmd())
	cmd.AddCommand(NewStatusCmd())
	cmd.AddCommand(NewListCmd())

	return cmd
}

//nolint:gochecknoglobals // Injected for testability to simulate help failures.
var helpRunner = func(cmd *cobra.Command) error {
	return cmd.Help()
}

func handleClusterRunE(cmd *cobra.Command, _ []string) error {
	// Cobra Help() can return an error (e.g., output stream or template issues); wrap it for clarity.
	err := helpRunner(cmd)
	if err != nil {
		return fmt.Errorf("displaying cluster command help: %w", err)
	}

	return nil
}
