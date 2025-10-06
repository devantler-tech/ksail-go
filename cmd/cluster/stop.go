package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/runtime"
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "stop",
		Short:        "Stop a running cluster",
		Long:         `Stop a running Kubernetes cluster.`,
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(rt)

	return cmd
}
