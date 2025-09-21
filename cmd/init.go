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

	return cmdhelpers.NewCobraCommand(
		"init",
		"Initialize a new project",
		"Initialize a new project.",
		func(cmd *cobra.Command, configManager *configmanager.ConfigManager, args []string) error {
			return HandleInitRunE(cmd, configManager, args)
		},
		fieldSelectors...,
	)
}

// HandleInitRunE handles the init command with an optional output path.
// If outputPath is empty, uses the current working directory.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	args []string,
	outputPath ...string,
) error {
	cluster, err := configManager.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load cluster config: %w", err)
	}

	// Determine output path
	var targetPath string
	if len(outputPath) > 0 && outputPath[0] != "" {
		targetPath = outputPath[0]
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
