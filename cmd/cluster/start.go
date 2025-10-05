package cluster

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		helpers.HandleConfigLoadRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}
