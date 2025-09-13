// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	return ksail.NewCobraCommand(
		"init",
		"Initialize a new project",
		`Initialize a new project.`,
		HandleInitRunE,
		ksail.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to use",
				&c.Spec.SourceDirectory, "k8s", "Directory containing workloads to deploy",
			}
		})...,
	)
}

// HandleInitRunE handles the init command.
// Exported for testing purposes.
func HandleInitRunE(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
	_ []string,
) error {
	cluster, err := utils.LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return err
	}

	notify.Successln(cmd.OutOrStdout(),
		"project initialized successfully")
	utils.LogClusterInfo(cmd, []utils.ClusterInfoField{
		{"Distribution", string(cluster.Spec.Distribution)},
		{"Source directory", cluster.Spec.SourceDirectory},
	})

	return nil
}
