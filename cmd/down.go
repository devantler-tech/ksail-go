// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to destroy",
			DefaultValue: v1alpha1.DistributionKind,
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to destroy",
			DefaultValue: "kind-ksail-default",
		},
	}

	// Create configuration manager with field selectors
	configManager := ksail.NewManager(fieldSelectors...)

	// Create the command
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Destroy a cluster",
		Long:  `Destroy a cluster.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := utils.HandleSimpleClusterCommand(
				cmd,
				configManager,
				"cluster destroyed successfully",
			)
			if err != nil {
				return fmt.Errorf("failed to handle cluster command: %w", err)
			}

			return nil
		},
	}

	// Add flags for the field selectors
	configManager.AddFlagsFromFields(cmd)

	return cmd
}
