// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// ErrInvalidConfigManagerType is returned when config manager type assertion fails.
var ErrInvalidConfigManagerType = errors.New("invalid config manager type")

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to list clusters for",
			DefaultValue: v1alpha1.DistributionKind,
		},
	}

	// Create the command using the helper
	cmd := cmdhelpers.NewCobraCommand(
		"list",
		"List clusters",
		`List all Kubernetes clusters managed by KSail.`,
		HandleListRunE,
		fieldSelectors...,
	)

	// Add the special --all flag manually since it's CLI-only
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")

	return cmd
}

// HandleListRunE handles the list command.
// Exported for testing purposes.
func HandleListRunE(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
	_ []string,
) error {
	// Type assert to concrete type to access exported Viper field
	ksailManager, ok := configManager.(*ksail.ConfigManager)
	if !ok {
		return ErrInvalidConfigManagerType
	}

	// Bind the --all flag manually since it's added after command creation
	_ = ksailManager.Viper.BindPFlag("all", cmd.Flags().Lookup("all"))

	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadConfig()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	all := ksailManager.Viper.GetBool("all")
	if all {
		notify.Successln(cmd.OutOrStdout(), "Listing all clusters (stub implementation)")
	} else {
		notify.Successln(cmd.OutOrStdout(), "Listing running clusters (stub implementation)")
	}

	notify.Activityln(cmd.OutOrStdout(),
		"Distribution filter: "+string(cluster.Spec.Distribution))

	return nil
}
