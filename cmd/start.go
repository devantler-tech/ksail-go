// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return utils.NewSimpleClusterCommand(utils.CommandConfig{
		Use:   "start",
		Short: "Start a stopped cluster",
		Long:  `Start a previously stopped cluster.`,
		RunEFunc: func(cmd *cobra.Command, configManager configmanager.ConfigManager[v1alpha1.Cluster], _ []string) error {
			_, err := utils.HandleSimpleClusterCommand(
				cmd,
				configManager,
				"Cluster started successfully (stub implementation)",
			)

			return err
		},
		FieldsFunc: func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to start",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context of cluster to start",
			}
		},
	})
}
