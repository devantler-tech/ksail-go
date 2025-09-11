// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"time"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	return config.NewCobraCommand(
		"status",
		"Show status of the Kubernetes cluster",
		`Show the current status of the Kubernetes cluster.`,
		handleStatusRunE,
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context to check status for",
				&c.Spec.Connection.Kubeconfig, "~/.kube/config", "Path to kubeconfig file",
				&c.Spec.Connection.Timeout,
				metav1.Duration{Duration: defaultStatusTimeout},
				"Timeout for status check operations",
			}
		})...,
	)
}

// handleStatusRunE handles the status command.
func handleStatusRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	cluster, err := loadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return err
	}

	notify.Successln(cmd.OutOrStdout(), "Cluster status: Running (stub implementation)")
	logClusterInfo(cmd, []ClusterInfoField{
		{"Context", cluster.Spec.Connection.Context},
		{"Kubeconfig", cluster.Spec.Connection.Kubeconfig},
	})

	return nil
}
