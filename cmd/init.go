// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new KSail project",
		Long:  `Initialize a new KSail project with the specified configuration options.`,
		RunE:  handleInitRunE,
	}

	// Add flags for initialization options
	cmd.Flags().String("container-engine", "Docker", "Container engine to use (Docker, Podman)")
	cmd.Flags().String("distribution", "Kind", "Kubernetes distribution to use (Kind, K3d)")

	return cmd
}

// handleInitRunE handles the init command.
func handleInitRunE(cmd *cobra.Command, _ []string) error {
	notify.Successln(cmd.OutOrStdout(), "Project initialized successfully (stub implementation)")

	return nil
}