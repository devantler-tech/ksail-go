package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewDeleteCmd creates and returns the delete command.
func NewDeleteCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Destroy a cluster",
		Long:         `Destroy a cluster.`,
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(runtimeContainer)

	return cmd
}
