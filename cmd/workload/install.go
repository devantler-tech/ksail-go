package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "Install Helm charts",
		Long:         "Install Helm charts to provision workloads through KSail.",
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(runtimeContainer)

	return cmd
}
