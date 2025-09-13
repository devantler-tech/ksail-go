// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return utils.NewSimpleClusterCommand(utils.CommandConfig{
		Use:   "stop",
		Short: "Stop the Kubernetes cluster",
		Long:  `Stop the Kubernetes cluster without removing it.`,
		RunEFunc: func(cmd *cobra.Command, configManager configmanager.ConfigManager[v1alpha1.Cluster], _ []string) error {
			_, err := utils.HandleSimpleClusterCommand(
				cmd,
				configManager,
				"Cluster stopped successfully (stub implementation)",
			)

			return err
		},
		FieldsFunc: func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to stop",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context of cluster to stop",
			}
		},
	})
}
