// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return NewSimpleClusterCommand(CommandConfig{
		Use:   "start",
		Short: "Start a stopped cluster",
		Long:  `Start a previously stopped cluster.`,
		RunEFunc: func(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
			_, err := HandleSimpleClusterCommand(
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
