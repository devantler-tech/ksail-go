// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultUpTimeout = 5 * time.Minute

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	// Create field selectors
	fieldSelectors := []ksail.FieldSelector[v1alpha1.Cluster]{
		cmdhelpers.StandardDistributionFieldSelector("Kubernetes distribution to use"),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			Description:  "Kubernetes context to use",
			DefaultValue: "kind-ksail-default",
		},
		{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for cluster operations",
			DefaultValue: metav1.Duration{Duration: defaultUpTimeout},
		},
	}

	// Create configuration manager with field selectors
	configManager := ksail.NewManager(fieldSelectors...)

	// Create the command
	//nolint:exhaustruct // Cobra commands intentionally use only required fields
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Start the Kubernetes cluster",
		Long:  `Start the Kubernetes cluster defined in the project configuration.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := cmdhelpers.HandleSimpleClusterCommand(
				cmd,
				configManager,
				"Cluster created and started successfully (stub implementation)",
			)
			if err != nil {
				return fmt.Errorf("failed to handle cluster command: %w", err)
			}

			return nil
		},
	}

	// Add flags for the field selectors
	configManager.AddFlagsFromFields(cmd)

	return cmd
}
