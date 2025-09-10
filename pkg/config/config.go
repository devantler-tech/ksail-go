// Package config provides centralized configuration management using Viper.
package config

import (
	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
)

// NewCobraCommand creates a cobra.Command with automatic type-safe configuration binding.
// This is the only constructor provided for initializing CobraCommands.
// The binding automatically handles CLI flags (priority 1), environment variables (priority 2),
// and configuration defaults (priority 3).
//
// If fieldSelectors is provided, only those specific fields will be bound as CLI flags.
// Field selectors can include optional descriptions.
// If fieldSelectors is empty, no configuration flags will be added (no auto-discovery by default).
//
// Usage examples:
//
//	// No configuration flags (default behavior):
//	config.NewCobraCommand("status", "Show status", "...", handleStatusRunE)
//
//	// Type-safe selective binding with direct field pointers (zero maintenance):
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//	        return []any{&c.Spec.Distribution, &c.Spec.SourceDirectory}
//	    })...)
//
//	// With embedded descriptions using AddFlagsFromFields:
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
//	        return []any{
//	            &c.Spec.Distribution, "Kubernetes distribution to use (EKS, K3d, Kind [default], Tind)",
//	            &c.Spec.SourceDirectory, "Directory containing workloads to deploy",
//	        }
//	    })...)
//
//	// Individual field selectors with descriptions:
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }, 
//	        "Kubernetes distribution to use (EKS, K3d, Kind [default], Tind)"),
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory }, 
//	        "Directory containing workloads to deploy"))
//
//	// Mixed approach - some fields with descriptions, others without:
//	config.NewCobraCommand("init", "Initialize", "...", handleInitRunE,
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution }),
//	    config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory }, 
//	        "Directory containing workloads to deploy"))
func NewCobraCommand(
	use, short, long string,
	runE func(*cobra.Command, *Manager, []string) error,
	fieldSelectors ...FieldSelector[v1alpha1.Cluster],
) *cobra.Command {
	manager := NewManager()

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
		// Bind only the specified field selectors
		bindFieldSelectors(cmd, manager, fieldSelectors)
	}
	// No else clause - when no field selectors provided, no configuration flags are added

	return cmd
}


