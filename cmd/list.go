// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	cmd := config.NewCobraCommand(
		"list",
		"List clusters",
		`List all Kubernetes clusters managed by KSail.`,
		handleListRunE,
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to list clusters for",
			}
		})...,
	)

	// Add the special --all flag manually since it's CLI-only
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones (default false)")

	return cmd
}

// handleListRunE handles the list command.
func handleListRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Bind the --all flag manually since it's added after command creation
	_ = configManager.GetViper().BindPFlag("all", cmd.Flags().Lookup("all"))

	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())
		return err
	}

	all := configManager.GetViper().GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters (stub implementation)")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters (stub implementation)")
	}

	notify.Activityln(cmd.OutOrStdout(),
		"Distribution filter: "+string(cluster.Spec.Distribution))

	return nil
}
