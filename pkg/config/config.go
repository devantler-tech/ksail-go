// Package config provides centralized configuration management using Viper.
package config

import (
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
)

// NewCobraCommand creates a cobra.Command with automatic type-safe configuration binding.
// This is the only constructor provided for initializing CobraCommands.
// The binding automatically handles CLI flags (priority 1), environment variables (priority 2),
// configuration files (priority 3), and field selector defaults (priority 4).
//
// If fieldSelectors is provided, only those specific fields will be bound as CLI flags.
// Field selectors must include default values and optionally descriptions.
// If fieldSelectors is empty, no configuration flags will be added (no auto-discovery by default).
//
// Usage examples:
//
//	// No configuration flags (default behavior):
//	config.NewCobraCommand("status", "Show status", "...", handleStatusRunE)
//
//	// Type-safe selective binding with defaults and descriptions:
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//	        return []any{
//	            &c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to use",
//	            &c.Spec.SourceDirectory, "k8s", "Directory containing workloads to deploy",
//	        }
//	    })...)
//
//	// With defaults only (description inferred):
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//	        return []any{
//	            &c.Spec.Distribution, v1alpha1.DistributionKind,
//	            &c.Spec.SourceDirectory, "k8s",
//	        }
//	    })...)
//
//	// Individual field selectors with defaults and descriptions:
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//	        v1alpha1.DistributionKind, "Kubernetes distribution to use"),
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
//	        "k8s", "Directory containing workloads to deploy"))
//
//	// Individual field selectors with defaults only:
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
//	        v1alpha1.DistributionKind),
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
//	        "k8s"))
func NewCobraCommand(
	use, short, long string,
	runE func(*cobra.Command, *Manager, []string) error,
	fieldSelectors ...FieldSelector[v1alpha1.Cluster],
) *cobra.Command {
	manager := NewManagerWithFieldSelectors(fieldSelectors)

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
		bindFieldSelectors(cmd, manager, fieldSelectors)
	}
	// No else clause - when no field selectors provided, no configuration flags are added

	return cmd
}
