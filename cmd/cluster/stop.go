package cluster

import (
	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		helpers.HandleConfigLoadRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}
