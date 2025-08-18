// Package inputs provides centralized flag management for KSail CLI commands.
package inputs

import "github.com/spf13/cobra"

// AddInitFlags adds all initialization-related flags to the given command.
func AddInitFlags(cmd *cobra.Command) {
	cmd.Flags().String("container-engine", "Docker", "Container engine to use (Docker, Podman)")
	cmd.Flags().String("distribution", "Kind", "Kubernetes distribution to use (Kind, K3d)")
	cmd.Flags().String("deployment-tool", "Kubectl", "Deployment tool to use (Kubectl, Flux)")
	cmd.Flags().String("secret-manager", "", "Secret manager to use (SOPS)")
	cmd.Flags().String("cni", "", "CNI to use (Default, Cilium, None)")
	cmd.Flags().String("csi", "", "CSI to use (Default, LocalPathProvisioner, None)")
	cmd.Flags().String("ingress-controller", "", "Ingress controller to use (Default, Traefik, None)")
	cmd.Flags().String("gateway-controller", "", "Gateway controller to use (Default, None)")
	cmd.Flags().String("metrics-server", "", "Enable metrics server (True, False)")
	cmd.Flags().String("mirror-registries", "", "Enable mirror registries (True, False)")
}

// AddListFlags adds list-related flags to the given command.
func AddListFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("all", false, "List all clusters including stopped ones")
}