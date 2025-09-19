// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/inputs"
	"github.com/devantler-tech/ksail-go/internal/scaffolder"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
)

// NewInitCmd creates and returns the init command.
func NewInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Scaffold a new project",
		Long: `Scaffold a new Kubernetes project in the specified directory.

  Includes:
  - 'ksail.yaml' configuration file for configuring KSail
  - 'kind.yaml'|'k3d.yaml'|'talos/*' configuration file(s) for configuring the distribution
  - '.sops.yaml' for managing secrets with SOPS (optional)
  - 'k8s/kustomization.yaml' as an entry point for Kustomize
  `,
		RunE: handleInitRunE,
	}

	// Add flags
	inputs.AddNameFlag(cmd)
	inputs.AddDistributionFlag(cmd)
	inputs.AddReconciliationToolFlag(cmd)
	inputs.AddSourceDirectoryFlag(cmd)
	inputs.AddForceFlag(cmd, "overwrite files")

	return cmd
}

// --- internals ---

// handleInitRunE handles the init command.
func handleInitRunE(cmd *cobra.Command, args []string) error {
	cfg := v1alpha1.NewCluster()
	
	// Read inputs from command flags
	if err := inputs.SetInputsFromCommand(cmd, cfg); err != nil {
		return fmt.Errorf("failed to process command inputs: %w", err)
	}

	return scaffold(cmd, cfg)
}

// scaffold generates initial project files according to the provided configuration.
func scaffold(cmd *cobra.Command, cfg *v1alpha1.Cluster) error {
	scaffolder := scaffolder.NewScaffolder(*cfg)

	cmd.Println("Scaffolding new project")

	err := scaffolder.Scaffold(inputs.Output, inputs.Force)
	if err != nil {
		return err
	}

	cmd.Println("✔ project scaffolded")

	return nil
}
