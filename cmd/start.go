// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return config.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		handleStartRunE,
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, "Kubernetes distribution to start",
				&c.Spec.Connection.Context, "Kubernetes context of cluster to start",
			}
		})...,
	)
}

// handleStartRunE handles the start command.
func handleStartRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())
		return err
	}

	notify.Successln(cmd.OutOrStdout(), "Cluster started successfully (stub implementation)")
	notify.Activityln(cmd.OutOrStdout(),
		"Distribution: "+string(cluster.Spec.Distribution))
	notify.Activityln(cmd.OutOrStdout(),
		"Context: "+cluster.Spec.Connection.Context)

	return nil
}
