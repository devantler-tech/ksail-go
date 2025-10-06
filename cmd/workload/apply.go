package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewApplyCmd creates the workload apply command.
func NewApplyCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "Apply manifests",
		Long:         "Apply local Kubernetes manifests to your cluster.",
		SilenceUsage: true,
	}

	cmd.RunE = shared.NewConfigLoaderRunE(runtimeContainer)

	return cmd
}
