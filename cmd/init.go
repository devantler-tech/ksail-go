// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/factory"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	return factory.NewCobraCommandWithFlags(
		"init",
		"Initialize a new KSail project",
		`Initialize a new KSail project with the specified configuration options.`,
		handleInitRunE,
		func(cmd *cobra.Command) {
			cmd.Flags().String("distribution", "Kind", "Kubernetes distribution to use (Kind, K3d, EKS)")
		},
	)
}

// handleInitRunE handles the init command.
func handleInitRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Project initialized successfully (stub implementation)")

	return nil
}
