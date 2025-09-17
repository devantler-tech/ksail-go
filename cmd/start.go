// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		func(cmd *cobra.Command, manager *configmanager.ConfigManager, args []string) error {
			return cmdhelpers.StandardClusterCommandRunE(
				"Cluster started successfully (stub implementation)",
			)(cmd, manager, args)
		},
		cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to start"),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to start",
			DefaultValue: "kind-ksail-default",
		},
	)
}
