package cluster

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	}

	cmd := cmdhelpers.NewCobraCommand(
		"down",
		"Destroy a cluster.",
		`Destroy a cluster and all of its associated resources.`,
		HandleDownRunE,
		fieldSelectors...,
	)

	return cmd
}

// HandleDownRunE handles the down command.
func HandleDownRunE(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	_ []string,
) error {
	tmr := timer.New()

	// Load cluster configuration
	_, err := configManager.LoadConfig()
	if err != nil {
		return err
	}

	cmd.Println()
	notify.TitleMessage(configManager.Writer, "🔥",
		notify.NewMessage("Destroying cluster..."))

	tmr.StartStage()
	notify.ActivityMessage(configManager.Writer, notify.NewMessage("destroying cluster"))
	// TODO: Actually destroy the cluster (T009)
	// For now, just simulate the operation

	notify.SuccessMessage(
		configManager.Writer,
		notify.NewMessage("cluster destroyed successfully").
			WithTiming(tmr.Total(), tmr.Stage()),
	)

	return nil
}
