// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultUpTimeout = 5 * time.Minute

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	return config.NewCobraCommand(
		"up",
		"Start the Kubernetes cluster",
		`Start the Kubernetes cluster defined in the project configuration.`,
		handleUpRunE,
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Kubernetes distribution to use",
				&c.Spec.DistributionConfig, "kind.yaml", "Configuration file for the distribution",
				&c.Spec.Connection.Context, "kind-ksail-default", "Kubernetes context to use",
				&c.Spec.Connection.Timeout,
				metav1.Duration{Duration: defaultUpTimeout},
				"Timeout for cluster operations",
			}
		})...,
	)
}

// handleUpRunE handles the up command.
func handleUpRunE(cmd *cobra.Command, configManager *config.Manager, _ []string) error {
	// Load the full cluster configuration (Viper handles all precedence automatically)
	cluster, err := configManager.LoadCluster()
	if err != nil {
		notify.Errorln(cmd.OutOrStdout(), "Failed to load cluster configuration: "+err.Error())

		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(
		cmd.OutOrStdout(),
		"Cluster created and started successfully (stub implementation)",
	)
	notify.Activityln(cmd.OutOrStdout(),
		"Distribution: "+string(cluster.Spec.Distribution))
	notify.Activityln(cmd.OutOrStdout(),
		"Context: "+cluster.Spec.Connection.Context)

	return nil
}
