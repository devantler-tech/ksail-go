// Package inputs provides centralized flag management for KSail CLI commands.
package inputs

import "github.com/spf13/cobra"

// AddInitFlags adds all initialization-related flags to the given command.
func AddInitFlags(cmd *cobra.Command) {
	cmd.Flags().String("container-engine", "Docker", "Container engine to use (Docker, Podman)")
	cmd.Flags().String("distribution", "Kind", "Kubernetes distribution to use (Kind, K3d)")
}

// AddListFlags adds list-related flags to the given command.
func AddListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
}