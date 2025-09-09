// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return config.NewCobraCommand(
		"start",
		"Start a stopped Kubernetes cluster",
		`Start a previously stopped Kubernetes cluster.`,
		handleStartRunE,
		[]config.FieldSelector[v1alpha1.Cluster]{}, // No specific configuration flags needed
	)
}

// handleStartRunE handles the start command.
func handleStartRunE(cmd *cobra.Command, _ *config.Manager, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster started successfully (stub implementation)")

	return nil
}
