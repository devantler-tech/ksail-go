package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		HandleStartRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
	)
}

// HandleStartRunE handles the start command.
func HandleStartRunE(
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
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Get timing and format
	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, false)

	notify.Successf(
		cmd.OutOrStdout(),
		"Cluster started successfully (stub implementation) %s",
		timingStr,
	)
	cmdhelpers.LogClusterInfo(cmd, []cmdhelpers.ClusterInfoField{
		{Label: "Distribution", Value: string(cluster.Spec.Distribution)},
		{Label: "Context", Value: cluster.Spec.Connection.Context},
	})

	return nil
}
