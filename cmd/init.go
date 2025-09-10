// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	return config.NewCobraCommand(
		"init",
		"Initialize a new project",
		`Initialize a new project.`,
		handleInitRunE,
		config.Fields(func(c *v1alpha1.Cluster) []any {
			return []any{&c.Spec.Distribution, &c.Spec.SourceDirectory}
		})...,
	)
}

// handleInitRunE handles the init command.
func handleInitRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return err
	}

	notify.Successln(cmd.OutOrStdout(),
		"project initialized successfully")
	notify.Activityln(cmd.OutOrStdout(),
		"Cluster name: "+cluster.Metadata.Name)
	notify.Activityln(cmd.OutOrStdout(),
		"Source directory: "+cluster.Spec.SourceDirectory)

	return nil
}
