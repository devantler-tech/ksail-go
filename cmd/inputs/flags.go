// Package inputs provides input handling and flag management for KSail commands.
package inputs

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
)

// Global variables for shared command inputs.
var (
	// Output specifies the output directory for generated files.
	Output string
	// Force indicates whether to overwrite existing files.
	Force bool
)

// AddNameFlag adds a name flag to the command.
func AddNameFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&Output, "output", "o", ".", "Output directory for scaffolded files")
}

// AddDistributionFlag adds a distribution flag to the command.
func AddDistributionFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(
		"distribution",
		"d",
		string(v1alpha1.DistributionKind),
		"Kubernetes distribution to use",
	)
}

// AddReconciliationToolFlag adds a reconciliation tool flag to the command.
func AddReconciliationToolFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("reconciliation-tool", "r", string(v1alpha1.ReconciliationToolKubectl), "Reconciliation tool to use")
}

// AddSourceDirectoryFlag adds a source directory flag to the command.
func AddSourceDirectoryFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("source-directory", "s", "k8s", "Directory containing workloads to deploy")
}

// AddForceFlag adds a force flag to the command with the given description.
func AddForceFlag(cmd *cobra.Command, description string) {
	cmd.Flags().BoolVarP(&Force, "force", "f", false, description)
}

// SetInputsOrFallback sets the input values from command flags or falls back to cluster configuration.
func SetInputsOrFallback(cfg *v1alpha1.Cluster) {
	// Note: In a real implementation, we would need to access the cobra command
	// to read flag values. For now, we'll set sensible defaults.

	if cfg.Spec.Distribution == "" {
		cfg.Spec.Distribution = v1alpha1.DistributionKind
	}

	if cfg.Spec.SourceDirectory == "" {
		cfg.Spec.SourceDirectory = "k8s"
	}

	if cfg.Spec.ReconciliationTool == "" {
		cfg.Spec.ReconciliationTool = v1alpha1.ReconciliationToolKubectl
	}

	if cfg.Spec.DistributionConfig == "" {
		// Set default distribution config based on distribution
		switch cfg.Spec.Distribution {
		case v1alpha1.DistributionKind:
			cfg.Spec.DistributionConfig = "kind.yaml"
		case v1alpha1.DistributionK3d:
			cfg.Spec.DistributionConfig = "k3d.yaml"
		case v1alpha1.DistributionEKS:
			cfg.Spec.DistributionConfig = "eksctl.yaml"
		default:
			cfg.Spec.DistributionConfig = "kind.yaml"
		}
	}
}

// SetInputsFromCommand reads flag values from the cobra command and sets the cluster configuration.
func SetInputsFromCommand(cmd *cobra.Command, cfg *v1alpha1.Cluster) error {
	if err := readDistributionFlag(cmd, cfg); err != nil {
		return err
	}

	readSourceDirectoryFlag(cmd, cfg)

	if err := readReconciliationToolFlag(cmd, cfg); err != nil {
		return err
	}

	readOutputFlag(cmd)
	readForceFlag(cmd)

	// Apply fallback values
	SetInputsOrFallback(cfg)

	return nil
}

// readDistributionFlag reads the distribution flag from the command.
func readDistributionFlag(cmd *cobra.Command, cfg *v1alpha1.Cluster) error {
	if distributionFlag := cmd.Flags().Lookup("distribution"); distributionFlag != nil && distributionFlag.Changed {
		if err := cfg.Spec.Distribution.Set(distributionFlag.Value.String()); err != nil {
			return fmt.Errorf("invalid distribution: %w", err)
		}
	}

	return nil
}

// readSourceDirectoryFlag reads the source directory flag from the command.
func readSourceDirectoryFlag(cmd *cobra.Command, cfg *v1alpha1.Cluster) {
	if sourceDirFlag := cmd.Flags().Lookup("source-directory"); sourceDirFlag != nil && sourceDirFlag.Changed {
		cfg.Spec.SourceDirectory = sourceDirFlag.Value.String()
	}
}

// readReconciliationToolFlag reads the reconciliation tool flag from the command.
func readReconciliationToolFlag(cmd *cobra.Command, cfg *v1alpha1.Cluster) error {
	if reconcileFlag := cmd.Flags().Lookup("reconciliation-tool"); reconcileFlag != nil && reconcileFlag.Changed {
		if err := cfg.Spec.ReconciliationTool.Set(reconcileFlag.Value.String()); err != nil {
			return fmt.Errorf("invalid reconciliation tool: %w", err)
		}
	}

	return nil
}

// readOutputFlag reads the output flag from the command.
func readOutputFlag(cmd *cobra.Command) {
	if outputFlag := cmd.Flags().Lookup("output"); outputFlag != nil && outputFlag.Changed {
		Output = outputFlag.Value.String()
	}
}

// readForceFlag reads the force flag from the command.
func readForceFlag(cmd *cobra.Command) {
	if forceFlag := cmd.Flags().Lookup("force"); forceFlag != nil && forceFlag.Changed {
		if val := forceFlag.Value.String(); val == "true" {
			Force = true
		}
	}
}
