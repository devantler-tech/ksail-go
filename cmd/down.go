// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return ksail.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		func(cmd *cobra.Command, manager *ksail.Manager, _ []string) error {
			_, err := cmdhelpers.HandleSimpleClusterCommand(
				cmd,
				manager,
				"cluster destroyed successfully",
			)
			if err != nil {
				return fmt.Errorf("failed to handle cluster command: %w", err)
			}

			return nil
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			Description:  "Kubernetes distribution to destroy",
			DefaultValue: v1alpha1.DistributionKind,
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to destroy",
			DefaultValue: "kind-ksail-default",
		},
	)
}
