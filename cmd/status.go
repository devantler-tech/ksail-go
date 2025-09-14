// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"status",
		"Show status of the Kubernetes cluster",
		`Show the current status of the Kubernetes cluster.`,
		HandleStatusRunE,
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context to check status for",
			DefaultValue: "kind-ksail-default",
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			Description:  "Path to kubeconfig file",
			DefaultValue: "~/.kube/config",
		},
		ksail.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for status check operations",
			DefaultValue: metav1.Duration{Duration: defaultStatusTimeout},
		},
	)
}

// HandleStatusRunE handles the status command.
// Exported for testing purposes.
func HandleStatusRunE(
	cmd *cobra.Command,
	manager *ksail.ConfigManager,
	_ []string,
) error {
	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(cmd, manager)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), "Cluster status: Running (stub implementation)")
	cmdhelpers.LogClusterInfo(cmd, []cmdhelpers.ClusterInfoField{
		{Label: "Context", Value: cluster.Spec.Connection.Context},
		{Label: "Kubeconfig", Value: cluster.Spec.Connection.Kubeconfig},
	})

	return nil
}
