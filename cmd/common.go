// Package cmd provides the command-line interface for KSail.
package cmd

import "github.com/spf13/cobra"

const suggestionsMinimumDistance = 2

// NewCobraCommand creates a cobra.Command with all fields explicitly set to avoid exhaustruct linting issues.
func NewCobraCommand(use, short, long string, runE func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:                        use,
		Short:                      short,
		Long:                       long,
		RunE:                       runE,
		Aliases:                    nil,
		SuggestFor:                 nil,
		GroupID:                    "",
		Example:                    "",
		ValidArgs:                  nil,
		ValidArgsFunction:          nil,
		Args:                       nil,
		ArgAliases:                 nil,
		BashCompletionFunction:     "",
		Deprecated:                 "",
		Annotations:                nil,
		Version:                    "",
		PersistentPreRun:           nil,
		PersistentPreRunE:          nil,
		PreRun:                     nil,
		PreRunE:                    nil,
		Run:                        nil,
		PostRun:                    nil,
		PostRunE:                   nil,
		PersistentPostRun:          nil,
		PersistentPostRunE:         nil,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: false,
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd:   false,
			DisableNoDescFlag:   false,
			DisableDescriptions: false,
			HiddenDefaultCmd:    false,
		},
		TraverseChildren:           false,
		Hidden:                     false,
		SilenceErrors:              false,
		SilenceUsage:               false,
		DisableFlagParsing:         false,
		DisableAutoGenTag:          false,
		DisableFlagsInUseLine:      false,
		DisableSuggestions:         false,
		SuggestionsMinimumDistance: suggestionsMinimumDistance,
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
