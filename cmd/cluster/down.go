package cluster

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		helpers.HandleConfigLoadRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}
