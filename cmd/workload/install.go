package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/runtime"
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd(rt *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "Install Helm charts",
		Long:         "Install Helm charts to provision workloads through KSail.",
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(rt)

	return cmd
}
