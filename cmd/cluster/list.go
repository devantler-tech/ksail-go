package cluster

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewListCmd creates and returns the list command.
func NewListCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
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
	configManager *configmanager.ConfigManager,
	_ []string,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Bind the --all flag manually since it's added after command creation
	_ = configManager.Viper.BindPFlag("all", cmd.Flags().Lookup("all"))

	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadConfig()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	all := configManager.Viper.GetBool("all")

	// Get timing and format
	total, stage := tmr.GetTiming()
	timingStr := notify.FormatTiming(total, stage, false)

	if all {
		notify.Successf(
			cmd.OutOrStdout(),
			"Listing all clusters (stub implementation) %s",
			timingStr,
		)
	} else {
		notify.Successf(cmd.OutOrStdout(), "Listing running clusters (stub implementation) %s", timingStr)
	}

	notify.Activityln(cmd.OutOrStdout(),
		"Distribution filter: "+string(cluster.Spec.Distribution))

	return nil
}
