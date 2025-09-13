// Package cmd provides the command-line interface for KSail.
package cmd

import (
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
		HandleReconcileRunE,
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

// HandleReconcileRunE handles the reconcile command.
// Exported for testing purposes.
func HandleReconcileRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	cluster, err := LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return err
	}

	notify.Successln(cmd.OutOrStdout(), "Workloads reconciled successfully (stub implementation)")
	logClusterInfo(cmd, []ClusterInfoField{
		{"Reconciliation tool", string(cluster.Spec.ReconciliationTool)},
		{"Source directory", cluster.Spec.SourceDirectory},
		{"Context", cluster.Spec.Connection.Context},
	})

	return nil
}
