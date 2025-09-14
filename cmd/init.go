// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to use"),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}

	// Create configuration manager with field selectors
	configManager := ksail.NewConfigManager(fieldSelectors...)

	// Create the command
	cmd := &cobra.Command{
		Use:                    "init",
		Aliases:                nil,
		SuggestFor:             nil,
		Short:                  "Initialize a new project",
		GroupID:                "",
		Long:                   `Initialize a new project.`,
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
		RunE: func(cmd *cobra.Command, args []string) error {
			return HandleInitRunE(cmd, configManager, args)
		},
		PostRun:            nil,
		PostRunE:           nil,
		PersistentPostRun:  nil,
		PersistentPostRunE: nil,
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
		SuggestionsMinimumDistance: 0,
	}

	// Add flags for the field selectors
	configManager.AddFlagsFromFields(cmd)

	return cmd
}

// HandleInitRunE handles the init command.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
	_ []string,
) error {
	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(),
		"project initialized successfully")
	cmdhelpers.LogClusterInfo(cmd, []cmdhelpers.ClusterInfoField{
		{Label: "Distribution", Value: string(cluster.Spec.Distribution)},
		{Label: "Source directory", Value: cluster.Spec.SourceDirectory},
	})

	return nil
}
