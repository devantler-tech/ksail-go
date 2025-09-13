// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return utils.NewSimpleClusterCommand(utils.CommandConfig{
		Use:   "down",
		Short: "Destroy a cluster",
		Long:  `Destroy a cluster.`,
		RunEFunc: func(cmd *cobra.Command, configManager configmanager.ConfigManager[v1alpha1.Cluster], _ []string) error {
			_, err := utils.HandleSimpleClusterCommand(
				cmd,
				configManager,
				"cluster destroyed successfully",
			)

			return err
		},
		FieldsFunc: func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to destroy",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context of cluster to destroy",
			}
		},
	})
}
