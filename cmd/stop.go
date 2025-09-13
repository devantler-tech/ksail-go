// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return ksail.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		func(cmd *cobra.Command, manager *ksail.Manager, _ []string) error {
			_, err := cmdhelpers.HandleSimpleClusterCommand(
				cmd,
				manager,
				"Cluster stopped successfully (stub implementation)",
			)
			if err != nil {
				return fmt.Errorf("failed to handle cluster command: %w", err)
			}

			return nil
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to stop",
			DefaultValue: v1alpha1.DistributionKind,
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to stop",
			DefaultValue: "kind-ksail-default",
		},
	)
}
