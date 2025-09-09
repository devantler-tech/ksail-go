// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	return factory.NewCobraCommandWithAutoBinding(
		"init",
		"Initialize a new KSail project",
		`Initialize a new KSail project with the specified configuration options.`,
		handleInitRunE,
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
	
	// Use the final resolved values from the cluster configuration
	distribution := string(cluster.Spec.Distribution)
	
	notify.Successln(cmd.OutOrStdout(), 
		"Project initialized successfully with "+distribution+" distribution (stub implementation)")
	notify.Activityln(cmd.OutOrStdout(), 
		"Cluster name: "+cluster.Metadata.Name)
	notify.Activityln(cmd.OutOrStdout(), 
		"Source directory: "+cluster.Spec.SourceDirectory)

	return nil
}
