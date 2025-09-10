// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return config.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		handleDownRunE,
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to destroy",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context of cluster to destroy",
			}
		})...,
	)
}

// handleDownRunE handles the down command.
func handleDownRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(
		cmd.OutOrStdout(),
		"cluster destroyed successfully",
	)
	notify.Activityln(cmd.OutOrStdout(),
		"Distribution: "+string(cluster.Spec.Distribution))
	notify.Activityln(cmd.OutOrStdout(),
		"Context: "+cluster.Spec.Connection.Context)

	return nil
}
