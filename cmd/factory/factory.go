// Package factory provides factory functions for creating CLI commands.
package factory

import (
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// SuggestionsMinimumDistance is the minimum edit distance to suggest a command.
const SuggestionsMinimumDistance = 2

// NewCobraCommand creates a cobra.Command.
func NewCobraCommand(use, short, long string, runE func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:                    use,
		Short:                  short,
		Long:                   long,
		RunE:                   runE,
		Aliases:                nil,
		SuggestFor:             nil,
		GroupID:                "",
		Example:                "",
		ValidArgs:              nil,
		ValidArgsFunction:      nil,
		Args:                   nil,
		ArgAliases:             nil,
		BashCompletionFunction: "",
		Deprecated:             "",
		Annotations:            nil,
		Version:                "",
		PersistentPreRun:       nil,
		PersistentPreRunE:      nil,
		PreRun:                 nil,
		PreRunE:                nil,
		Run:                    nil,
		PostRun:                nil,
		PostRunE:               nil,
		PersistentPostRun:      nil,
		PersistentPostRunE:     nil,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: false,
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:         false,
			DisableNoDescFlag:         false,
			DisableDescriptions:       false,
			HiddenDefaultCmd:          false,
			DefaultShellCompDirective: nil,
		},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: SuggestionsMinimumDistance,
	}
}

// NewCobraCommandWithFlags creates a new cobra.Command with complete field initialization and flag setup.
func NewCobraCommandWithFlags(
	use, short, long string,
	runE func(*cobra.Command, []string) error,
	setupFlags func(*cobra.Command),
) *cobra.Command {
	cmd := NewCobraCommand(use, short, long, runE)
	setupFlags(cmd)

	return cmd
}

// NewCobraCommandWithAutoBinding creates a new cobra.Command with automatic flag binding.
// This eliminates the need for manual flag setup and binding.
func NewCobraCommandWithAutoBinding(
	use, short, long string,
	runE func(*cobra.Command, *config.Manager, []string) error,
) *cobra.Command {
	configManager := config.NewManager()
	
	cmd := NewCobraCommand(use, short, long, func(cmd *cobra.Command, args []string) error {
		return runE(cmd, configManager, args)
	})
	
	// Automatically bind all available flags based on the v1alpha1.Cluster structure
	configManager.AutoBindFlags(cmd)
	
	return cmd
}

// NewCobraCommandWithSelectiveBinding creates a new cobra.Command with selective automatic flag binding.
// Only binds the specified flags instead of all available flags.
func NewCobraCommandWithSelectiveBinding(
	use, short, long string,
	runE func(*cobra.Command, *config.Manager, []string) error,
	flagNames []string,
) *cobra.Command {
	configManager := config.NewManager()
	
	cmd := NewCobraCommand(use, short, long, func(cmd *cobra.Command, args []string) error {
		return runE(cmd, configManager, args)
	})
	
	// Only bind specific flags instead of all flags
	if len(flagNames) > 0 {
		bindSelectiveFlags(cmd, configManager, flagNames)
	} else {
		// If no specific flags requested, bind all
		configManager.AutoBindFlags(cmd)
	}
	
	return cmd
}

// bindSelectiveFlags binds only the specified flags.
func bindSelectiveFlags(cmd *cobra.Command, configManager *config.Manager, flagNames []string) {
	// Create a map for quick lookup
	flagsToInclude := make(map[string]bool)
	for _, name := range flagNames {
		flagsToInclude[name] = true
	}
	
	// Bind specific flags through the config manager
	configManager.BindSelectiveFlags(cmd, flagsToInclude)
}
