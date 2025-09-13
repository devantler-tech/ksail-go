// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/utils"
	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultStatusTimeout = 5 * time.Minute

// NewStatusCmd creates and returns the status command.
func NewStatusCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context to check status for",
			DefaultValue: "kind-ksail-default",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			Description:  "Path to kubeconfig file",
			DefaultValue: "~/.kube/config",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for status check operations",
			DefaultValue: metav1.Duration{Duration: defaultStatusTimeout},
		},
	}

	// Create configuration manager with field selectors
	configManager := ksail.NewManager(fieldSelectors...)

	// Create the command
	//nolint:exhaustruct // Cobra commands intentionally use only required fields
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show status of the Kubernetes cluster",
		Long:  `Show the current status of the Kubernetes cluster.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return HandleStatusRunE(cmd, configManager, args)
		},
	}

	// Add flags for the field selectors
	configManager.AddFlagsFromFields(cmd)

	return cmd
}

// HandleStatusRunE handles the status command.
// Exported for testing purposes.
func HandleStatusRunE(
	cmd *cobra.Command,
	configManager configmanager.ConfigManager[v1alpha1.Cluster],
	_ []string,
) error {
	cluster, err := utils.LoadClusterWithErrorHandling(cmd, configManager)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	notify.Successln(cmd.OutOrStdout(), "Cluster status: Running (stub implementation)")
	utils.LogClusterInfo(cmd, []utils.ClusterInfoField{
		{Label: "Context", Value: cluster.Spec.Connection.Context},
		{Label: "Kubeconfig", Value: cluster.Spec.Connection.Kubeconfig},
	})

	return nil
}
