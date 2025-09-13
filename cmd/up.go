// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultUpTimeout = 5 * time.Minute

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	return utils.NewSimpleClusterCommand(utils.CommandConfig{
		Use:   "up",
		Short: "Start the Kubernetes cluster",
		Long:  `Start the Kubernetes cluster defined in the project configuration.`,
		RunEFunc: func(cmd *cobra.Command, configManager configmanager.ConfigManager[v1alpha1.Cluster], _ []string) error {
			_, err := utils.HandleSimpleClusterCommand(
				cmd,
				configManager,
				"Cluster created and started successfully (stub implementation)",
			)

			return err
		},
		FieldsFunc: func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to use",
				&c.Spec.DistributionConfig, "kind.yaml", "Configuration file for the distribution",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context to use",
				&c.Spec.Connection.Timeout,
				metav1.Duration{Duration: defaultUpTimeout},
				"Timeout for cluster operations",
			}
		},
	})
}
