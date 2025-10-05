package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/spf13/cobra"
)

// NewApplyCmd creates the workload apply command.
func NewApplyCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"apply",
		"Apply manifests",
		"Apply local Kubernetes manifests to your cluster.",
		helpers.HandleConfigLoadRunE,
	)
}
