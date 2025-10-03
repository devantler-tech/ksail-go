package cluster

import (
	"context"
	"fmt"
	"time"

	"github.com/devantler-tech/ksail-go/cmd/internal/cmdhelpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
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
	return handleUpRunEWithProvisioner(cmd, manager, nil)
}

// handleUpRunEWithProvisioner is the internal implementation that accepts an optional provisioner for testing.
func handleUpRunEWithProvisioner(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	provisioner provisionerFactory,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Load cluster configuration
	cluster, err := cmdhelpers.LoadClusterWithErrorHandling(cmd, manager, tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %w", err)
	}

	// Transition to provisioning stage
	tmr.NewStage()

	// Show provisioning title
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Create cluster...",
		Emoji:   "🚀",
		Writer:  cmd.OutOrStdout(),
	})

	// Create provisioner based on distribution
	var clusterProvisioner clusterprovisioner.ClusterProvisioner

	var clusterName string

	if provisioner != nil {
		clusterProvisioner, clusterName, err = provisioner(cmd.Context(), cluster)
	} else {
		// Load config once and get both provisioner and cluster name
		clusterProvisioner, clusterName, err = cmdhelpers.CreateClusterProvisionerWithName(cmd.Context(), cluster)
	}

	if err != nil {
		return fmt.Errorf("failed to create provisioner: %w", err)
	}

	// Show activity message
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "creating cluster",
		Writer:  cmd.OutOrStdout(),
	})

	// Provision the cluster
	err = clusterProvisioner.Create(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to provision cluster: %w", err)
	}

	// Display success with timing
	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster created",
		Timer:      tmr,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// provisionerFactory is a function type for creating cluster provisioners (for testing).
// Returns provisioner, cluster name, and error.
type provisionerFactory func(context.Context, *v1alpha1.Cluster) (clusterprovisioner.ClusterProvisioner, string, error)
