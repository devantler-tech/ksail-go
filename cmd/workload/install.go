package workload

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/spf13/cobra"
)

// NewInstallCmd creates the workload install command.
func NewInstallCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"install",
		"Install Helm charts",
		"Install Helm charts to provision workloads through KSail.",
		helpers.HandleConfigLoadRunE,
	)
}
