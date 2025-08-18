// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start a stopped Kubernetes cluster",
		Long:  `Start a previously stopped Kubernetes cluster.`,
		RunE:  handleStartRunE,
	}

	return cmd
}

// handleStartRunE handles the start command.
func handleStartRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Cluster started successfully (stub implementation)")

	return nil
}