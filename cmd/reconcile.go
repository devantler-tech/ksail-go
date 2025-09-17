// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanagerinterface "github.com/devantler-tech/ksail-go/pkg/config-manager"
	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates and returns the reconcile command.
func NewReconcileCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []configmanager.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.ReconciliationTool },
			Description:  "Tool to use for reconciling workloads",
			DefaultValue: v1alpha1.ReconciliationToolKubectl,
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			Description:  "Directory containing workloads to reconcile",
			DefaultValue: "k8s",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context to reconcile workloads in",
			DefaultValue: "kind-ksail-default",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			Description:  "Path to kubeconfig file",
			DefaultValue: "~/.kube/config",
		},
	}

	// Create the command using the helper
	return cmdhelpers.NewCobraCommand(
		"reconcile",
		"Reconcile workloads in the cluster",
		`Reconcile workloads in the cluster to match the desired state
defined in configuration files.`,
		HandleReconcileRunE,
		fieldSelectors...,
	)
}

// HandleReconcileRunE handles the reconcile command.
// Exported for testing purposes.
func HandleReconcileRunE(
	cmd *cobra.Command,
	configManager configmanagerinterface.ConfigManager[v1alpha1.Cluster],
	_ []string,
) error {
	err := cmdhelpers.ExecuteCommandWithClusterInfo(
		cmd,
		configManager,
		"Workloads reconciled successfully (stub implementation)",
		func(cluster *v1alpha1.Cluster) []cmdhelpers.ClusterInfoField {
			return []cmdhelpers.ClusterInfoField{
				{Label: "Reconciliation tool", Value: string(cluster.Spec.ReconciliationTool)},
				{Label: "Source directory", Value: cluster.Spec.SourceDirectory},
				{Label: "Context", Value: cluster.Spec.Connection.Context},
			}
		},
	)
	if err != nil {
		return fmt.Errorf("failed to execute reconcile command: %w", err)
	}

	return nil
}
