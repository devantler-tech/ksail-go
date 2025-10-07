package cluster

import (
	"fmt"

	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewClusterCmd creates the parent cluster command and wires lifecycle subcommands beneath it.
func NewClusterCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage cluster lifecycle",
		Long: `Manage lifecycle operations for local Kubernetes clusters, including ` +
			`provisioning, teardown, and status.`,
		RunE:         handleClusterRunE,
		SilenceUsage: true,
	}

	cmd.AddCommand(NewCreateCmd(runtimeContainer))
	cmd.AddCommand(NewDeleteCmd(runtimeContainer))
	cmd.AddCommand(NewStartCmd(runtimeContainer))
	cmd.AddCommand(NewStopCmd(runtimeContainer))
	cmd.AddCommand(NewStatusCmd(runtimeContainer))
	cmd.AddCommand(NewListCmd(runtimeContainer))

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
