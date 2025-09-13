// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
)

// NewReconcileCmd creates and returns the reconcile command.
func NewReconcileCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
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

	// Create configuration manager with field selectors
	configManager := ksail.NewManager(fieldSelectors...)

	// Create the command
	//nolint:exhaustruct // Cobra commands intentionally use only required fields
	cmd := &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile workloads in the cluster",
		Long: `Reconcile workloads in the cluster to match the desired state
defined in configuration files.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return HandleReconcileRunE(cmd, configManager, args)
		},
	}

	// Add flags for the field selectors
	configManager.AddFlagsFromFields(cmd)

	return cmd
}

// HandleReconcileRunE handles the reconcile command.
// Exported for testing purposes.
func HandleReconcileRunE(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
	_ []string,
) error {
	cluster, err := utils.LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), "Workloads reconciled successfully (stub implementation)")
	utils.LogClusterInfo(cmd, []utils.ClusterInfoField{
		{Label: "Reconciliation tool", Value: string(cluster.Spec.ReconciliationTool)},
		{Label: "Source directory", Value: cluster.Spec.SourceDirectory},
		{Label: "Context", Value: cluster.Spec.Connection.Context},
	})

	return nil
}
