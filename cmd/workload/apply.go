package workload

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/spf13/cobra"
)

// NewApplyCmd creates the workload apply command.
func NewApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply manifests",
		Long:  "Apply local Kubernetes manifests to your cluster.",
		RunE:  utils.HandleConfigLoadRunE,
	}
}
