// Package cmdhelpers provides common utilities for KSail command creation and handling.
package cmdhelpers

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/spf13/cobra"
)

// SuggestionsMinimumDistance is the minimum distance for command suggestions.
const SuggestionsMinimumDistance = 2

// NewCobraCommand creates a cobra.Command with automatic type-safe configuration binding
// for v1alpha1.Cluster configurations. This constructor provides a unified approach to
// CLI command creation with integrated configuration management.
//
// The configuration binding follows a priority hierarchy:
//
//  1. CLI flags (highest priority)
//  2. Environment variables
//  3. Configuration files (ksail.yaml)
//  4. Field selector defaults (lowest priority)
//
// Parameters:
//   - use: The command name and usage pattern
//   - short: Brief description shown in command list
//   - long: Detailed description shown in help
//   - runE: Command execution function with access to configuration manager
//   - fieldSelectors: Optional field selectors to expose as CLI flags
//
// When fieldSelectors are provided, only those specific fields are bound as CLI flags
// with type-safe validation and automatic help generation. When no fieldSelectors are
// provided, the command runs without configuration flags (suitable for status commands).
//
// Usage examples:
//
//	// Command without configuration flags:
//	NewCobraCommand("status", "Show cluster status", "...", handleStatusRunE)
//
//	// Command with type-safe configuration flags:
//	NewCobraCommand("init", "Initialize cluster", "...", handleInitRunE,
//	    AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//	        v1alpha1.DistributionKind, "Kubernetes distribution to use"),
//	    AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
//	        "k8s", "Directory containing workloads to deploy"))
func NewCobraCommand(
	use, short, long string,
	runE func(*cobra.Command, *configmanager.ConfigManager, []string) error,
	fieldSelectors ...configmanager.FieldSelector[v1alpha1.Cluster],
) *cobra.Command {
	// Create the base command first so we can access cmd.OutOrStdout()
	cmd := &cobra.Command{
		Use:                        use,
		Short:                      short,
		Long:                       long,
		SuggestionsMinimumDistance: SuggestionsMinimumDistance,
	}

	// Create the manager with the command's writer
	manager := configmanager.NewConfigManager(cmd.OutOrStdout(), fieldSelectors...)

	// Set the RunE function after manager is created
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return runE(cmd, manager, args)
	}

	// Auto-bind flags based on field selectors
	if len(fieldSelectors) > 0 {
		// Bind only the specified field selectors for CLI flags
		manager.AddFlagsFromFields(cmd)
	}
	// No else clause - when no field selectors provided, no configuration flags are added

	return cmd
}

// HandleSimpleClusterCommand provides common error handling and cluster loading for simple commands.
func HandleSimpleClusterCommand(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	successMessage string,
) (*v1alpha1.Cluster, error) {
	// Load the full cluster configuration with validation
	cluster, err := configManager.LoadConfig()
	if err != nil {
		return nil, err
	}

	notify.SuccessMessage(cmd.OutOrStdout(), notify.NewMessage(successMessage))

	return cluster, nil
}

// StandardClusterCommandRunE creates a standard run function for cluster commands.
// It handles the common pattern of calling HandleSimpleClusterCommand with a success message.
func StandardClusterCommandRunE(
	successMessage string,
) func(cmd *cobra.Command, manager *configmanager.ConfigManager, args []string) error {
	return func(
		cmd *cobra.Command,
		manager *configmanager.ConfigManager,
		_ []string,
	) error {
		_, err := HandleSimpleClusterCommand(cmd, manager, successMessage)
		if err != nil {
			return fmt.Errorf("failed to handle cluster command: %w", err)
		}

		return nil
	}
}

// StandardDistributionFieldSelector creates a standard field selector for distribution.
func StandardDistributionFieldSelector() configmanager.FieldSelector[v1alpha1.Cluster] {
	return configmanager.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		Description:  "Kubernetes distribution to use",
		DefaultValue: v1alpha1.DistributionKind,
	}
}

// StandardSourceDirectoryFieldSelector creates a standard field selector for source directory.
func StandardSourceDirectoryFieldSelector() configmanager.FieldSelector[v1alpha1.Cluster] {
	return configmanager.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
		Description:  "Directory containing workloads to deploy",
		DefaultValue: "k8s",
	}
}

// StandardDistributionConfigFieldSelector creates a standard field selector for distribution config.
func StandardDistributionConfigFieldSelector() configmanager.FieldSelector[v1alpha1.Cluster] {
	return configmanager.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
		Description:  "Configuration file for the distribution",
		DefaultValue: "kind.yaml",
	}
}

// StandardContextFieldSelector creates a standard field selector for kubernetes context.
func StandardContextFieldSelector() configmanager.FieldSelector[v1alpha1.Cluster] {
	return configmanager.FieldSelector[v1alpha1.Cluster]{
		Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
		Description:  "Kubernetes context of cluster",
		DefaultValue: "kind-kind",
	}
}
