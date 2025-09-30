package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		func(cmd *cobra.Command, manager *configmanager.ConfigManager, args []string) error {
			tmr := timer.New()

			// Load cluster configuration
			_, err := manager.LoadConfig()
			if err != nil {
				return err
			}

			// TODO: Actually stop the cluster (T009)
			// For now, just simulate the operation

			notify.SuccessMessage(
				manager.Writer,
				notify.NewMessage("cluster stopped successfully").
					WithTiming(tmr.Total(), tmr.Stage()),
			)

			return nil
		},
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}
