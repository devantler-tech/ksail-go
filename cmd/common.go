// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// CommandConfig holds the configuration for creating a command.
type CommandConfig struct {
	Use        string
	Short      string
	Long       string
	RunEFunc   func(cmd *cobra.Command, configManager *config.Manager, args []string) error
	FieldsFunc func(c *v1alpha1.Cluster) []any
}

// NewSimpleClusterCommand creates a new command with common cluster management pattern.
func NewSimpleClusterCommand(cfg CommandConfig) *cobra.Command {
	return config.NewCobraCommand(
		cfg.Use,
		cfg.Short,
		cfg.Long,
		cfg.RunEFunc,
		config.AddFlagsFromFields(cfg.FieldsFunc)...,
	)
}

// HandleSimpleClusterCommand provides common error handling and cluster loading for simple commands.
func HandleSimpleClusterCommand(
	cmd *cobra.Command,
	configManager *config.Manager,
	successMessage string,
) (*v1alpha1.Cluster, error) {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return nil, fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), successMessage)
	notify.Activityln(cmd.OutOrStdout(), "Distribution: "+string(cluster.Spec.Distribution))
	notify.Activityln(cmd.OutOrStdout(), "Context: "+cluster.Spec.Connection.Context)

	return cluster, nil
}
