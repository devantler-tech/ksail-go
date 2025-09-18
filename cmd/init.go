// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to use"),
		cmdhelpers.StandardSourceDirectoryFieldSelector(),
	}

	// Use the common command creation helper
	return cmdhelpers.NewCobraCommand(
		"init",
		"Initialize a new project",
		"Initialize a new project.",
		HandleInitRunE,
		fieldSelectors...,
	)
}

// HandleInitRunE handles the init command.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	configManager *configmanager.ConfigManager,
	_ []string,
) error {
	err := cmdhelpers.ExecuteCommandWithClusterInfo(
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
