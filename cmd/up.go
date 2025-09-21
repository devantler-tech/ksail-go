// Package cmd provides the command-line interface for KSail.
package cmd

import (
	"fmt"
	"time"

	configmanager "github.com/devantler-tech/ksail-go/cmd/config-manager"
	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const defaultUpTimeout = 5 * time.Minute

// NewUpCmd creates and returns the up command.
func NewUpCmd() *cobra.Command {
	return cmdhelpers.NewCobraCommand(
		"up",
		"Start the Kubernetes cluster",
		`Start the Kubernetes cluster defined in the project configuration.`,
		HandleUpRunE,
		cmdhelpers.StandardDistributionFieldSelector(),
		cmdhelpers.StandardDistributionConfigFieldSelector(),
		cmdhelpers.StandardContextFieldSelector(),
		configmanager.FieldSelector[v1alpha1.Cluster]{
			Selector:     func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			Description:  "Timeout for cluster operations",
			DefaultValue: metav1.Duration{Duration: defaultUpTimeout},
		},
	)
}

// HandleUpRunE handles the up command.
// Exported for testing purposes.
func HandleUpRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	_, err := cmdhelpers.HandleSimpleClusterCommand(
		cmd,
		manager,
		"Cluster created and started successfully (stub implementation)",
	)
	if err != nil {
		return fmt.Errorf("failed to handle cluster command: %w", err)
	}

	return nil
}
