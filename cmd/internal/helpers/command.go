// Package helpers provides common utilities for KSail command creation and handling.
package helpers

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
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

// HandleConfigLoadRunE provides a shared implementation for commands whose primary
// responsibility is to load the cluster configuration and exit without additional
// side effects. It centralizes timer handling and error wrapping to keep
// individual commands focused on their specific logic.
func HandleConfigLoadRunE(
	_ *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	tmr := timer.New()
	tmr.Start()

	_, err := manager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	return nil
}
