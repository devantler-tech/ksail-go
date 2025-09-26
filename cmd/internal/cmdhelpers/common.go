// Package cmdhelpers provides common utilities for KSail command creation and handling.
package cmdhelpers

import (
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/helpers"
	ksailvalidator "github.com/devantler-tech/ksail-go/pkg/validator/ksail"
	"github.com/spf13/cobra"
)

// ClusterInfoField represents a field to log from cluster information.
type ClusterInfoField struct {
	Label string
	Value string
}

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
	manager := configmanager.NewConfigManager(fieldSelectors...)

	// Create the base command
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runE(cmd, manager, args)
		},
		SuggestionsMinimumDistance: SuggestionsMinimumDistance,
	}

	// Auto-bind flags based on field selectors
	if len(fieldSelectors) > 0 {
		// Bind only the specified field selectors for CLI flags
		manager.AddFlagsFromFields(cmd)
	}
	// No else clause - when no field selectors provided, no configuration flags are added

	return cmd
}

// LogClusterInfo logs cluster information fields to the command output.
func LogClusterInfo(cmd *cobra.Command, fields []ClusterInfoField) {
	for _, field := range fields {
		notify.Activityln(cmd.OutOrStdout(), field.Label+": "+field.Value)
	}
}

// LoadClusterWithErrorHandling provides common error handling pattern for loading cluster configuration.
// Exported for testing purposes.
func LoadClusterWithErrorHandling(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
) (*v1alpha1.Cluster, error) {
	cluster, err := configManager.LoadConfig()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return nil, fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Validate the loaded configuration
	validator := ksailvalidator.NewValidator()
	result := validator.Validate(cluster)

	// Handle validation errors with fail-fast behavior
	if !result.Valid {
		// Use standardized error formatting from helpers
		errorMessages := helpers.FormatValidationErrorsMultiline(result)
		notify.Errorln(cmd.OutOrStdout(),
			"Configuration validation failed:\n"+errorMessages)

		// Print fix suggestions using standardized helper
		fixSuggestions := helpers.FormatValidationFixSuggestions(result)
		for _, suggestion := range fixSuggestions {
			notify.Activityln(cmd.OutOrStdout(), suggestion)
		}

		// Display warnings using standardized helper
		warnings := helpers.FormatValidationWarnings(result)
		for _, warning := range warnings {
			notify.Warnln(cmd.OutOrStdout(), warning)
		}

		return nil, fmt.Errorf("%w with %d errors",
			helpers.ErrConfigurationValidationFailed, len(result.Errors))
	}

	// Display warnings even for valid configurations using standardized helper
	warnings := helpers.FormatValidationWarnings(result)
	for _, warning := range warnings {
		notify.Warnln(cmd.OutOrStdout(), warning)
	}

	return cluster, nil
}

// HandleSimpleClusterCommand provides common error handling and cluster loading for simple commands.
func HandleSimpleClusterCommand(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	successMessage string,
) (*v1alpha1.Cluster, error) {
	// Load the full cluster configuration using common error handling
	cluster, err := LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return nil, err
	}

	notify.Successln(cmd.OutOrStdout(), successMessage)
	LogClusterInfo(cmd, []ClusterInfoField{
		{"Distribution", string(cluster.Spec.Distribution)},
		{"Context", cluster.Spec.Connection.Context},
	})

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

// ExecuteCommandWithClusterInfo loads cluster configuration and executes a command with cluster info logging.
func ExecuteCommandWithClusterInfo(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	successMessage string,
	infoFieldsFunc func(*v1alpha1.Cluster) []ClusterInfoField,
) error {
	cluster, err := LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), successMessage)
	LogClusterInfo(cmd, infoFieldsFunc(cluster))

	return nil
}
