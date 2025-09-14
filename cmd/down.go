// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		func(cmd *cobra.Command, manager *ksail.ConfigManager, args []string) error {
			return cmdhelpers.StandardClusterCommandRunE(
				"cluster destroyed successfully",
			)(cmd, manager, args)
		},
		cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to destroy"),
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to destroy",
			DefaultValue: "kind-ksail-default",
		},
	)
}
