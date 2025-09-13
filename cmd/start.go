// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return ksail.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		func(cmd *cobra.Command, manager *ksail.Manager, _ []string) error {
			_, err := utils.HandleSimpleClusterCommand(
				cmd,
				manager,
				"Cluster started successfully (stub implementation)",
			)
			if err != nil {
				return fmt.Errorf("failed to handle cluster command: %w", err)
			}

			return nil
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to start",
			DefaultValue: v1alpha1.DistributionKind,
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to start",
			DefaultValue: "kind-ksail-default",
		},
	)
}
