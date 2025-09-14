// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		func(cmd *cobra.Command, manager configmanager.ConfigManager[v1alpha1.Cluster], args []string) error {
			return cmdhelpers.StandardClusterCommandRunE(
				"Cluster stopped successfully (stub implementation)",
			)(cmd, manager, args)
		},
		cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to stop"),
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context of cluster to stop",
			DefaultValue: "kind-ksail-default",
		},
	)
}
