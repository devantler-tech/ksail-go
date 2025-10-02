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
		HandleStopRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleStopRunE handles the stop command.
func HandleStopRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Load cluster and execute
	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(cmd, manager)
	if err != nil {
		return err
	}

	// Get timing and format
	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, false)

	notify.Successf(
		cmd.OutOrStdout(),
		"Cluster stopped successfully (stub implementation) %s",
		timingStr,
	)
	cmdhelpers.LogClusterInfo(cmd, []cmdhelpers.ClusterInfoField{
		{Label: "Distribution", Value: string(cluster.Spec.Distribution)},
		{Label: "Context", Value: cluster.Spec.Connection.Context},
	})

	return nil
}
