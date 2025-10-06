package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/runtime"
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start a stopped cluster",
		Long:         `Start a previously stopped cluster.`,
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(rt)

	return cmd
}
