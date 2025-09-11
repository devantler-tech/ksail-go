// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates and returns the reconcile command.
func NewReconcileCmd() *cobra.Command {
	return config.NewCobraCommand(
		"reconcile",
		"Reconcile workloads in the cluster",
		`Reconcile workloads in the cluster to match the desired state
defined in configuration files.`,
		handleReconcileRunE,
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.ReconciliationTool, v1alpha1.ReconciliationToolKubectl, "Tool to use for reconciling workloads",
				&c.Spec.SourceDirectory, "k8s", "Directory containing workloads to reconcile",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context to reconcile workloads in",
				&c.Spec.Connection.Kubeconfig, "~/.kube/config", "Path to kubeconfig file",
			}
		})...,
	)
}

// handleReconcileRunE handles the reconcile command.
func handleReconcileRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), "Workloads reconciled successfully (stub implementation)")
	logClusterInfo(cmd, []ClusterInfoField{
		{"Reconciliation tool", string(cluster.Spec.ReconciliationTool)},
		{"Source directory", cluster.Spec.SourceDirectory},
		{"Context", cluster.Spec.Connection.Context},
	})

	return nil
}
