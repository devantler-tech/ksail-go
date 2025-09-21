// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"
	"os"

	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardNameFieldSelector(),
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}

	// Create the command using the helper
	cmd := cmdhelpers.NewCobraCommand(
		"init",
		"Initialize a new project",
		"Initialize a new project.",
		HandleInitRunE,
		fieldSelectors...,
	)

	// Add the --output flag for specifying output directory
	cmd.Flags().StringP("output", "o", "", "Output directory for the project")

	return cmd
}

// HandleInitRunE handles the init command with an optional output path.
// If outputPath is empty, uses the current working directory.
// The variadic outputPath parameter is for testing purposes only.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	_ []string,
) error {
	// Bind the --output flag
	_ = configManager.Viper.BindPFlag("output", cmd.Flags().Lookup("output"))

	cluster, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load cluster config: %w", err)
	}

	// Determine target path - prioritize test parameter over flag
	var targetPath string
	// Get output path from flag
	flagOutputPath := configManager.Viper.GetString("output")
	if flagOutputPath != "" {
		targetPath = flagOutputPath
	} else {
		targetPath, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Create scaffolder and generate project files
	scaffolderInstance := scaffolder.NewScaffolder(*cluster)

	// Use targetPath directly - scaffolder will handle path joining
	err = scaffolderInstance.Scaffold(targetPath, false)
	if err != nil {
		return fmt.Errorf("failed to scaffold project files: %w", err)
	}

	err = cmdhelpers.ExecuteCommandWithClusterInfo(
		cmd,
		configManager,
		"project initialized successfully",
		func(cluster *v1alpha1.Cluster) []cmdhelpers.ClusterInfoField {
			return []cmdhelpers.ClusterInfoField{
				{Label: "Distribution", Value: string(cluster.Spec.Distribution)},
				{Label: "Source directory", Value: cluster.Spec.SourceDirectory},
			}
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute init command: %w", err)
	}

	return nil
}
