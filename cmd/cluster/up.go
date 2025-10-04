package cluster

import (
	"context"
	"fmt"
	"time"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
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
	return helpers.NewCobraCommand(
		"up",
		"Start the Kubernetes cluster",
		`Start the Kubernetes cluster defined in the project configuration.`,
		HandleUpRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardDistributionConfigFieldSelector(),
		configmanager.StandardContextFieldSelector(),
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
	cluster, err := manager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster: %w", err)
	}

	// Transition to provisioning stage
	tmr.NewStage()

	// Show provisioning title
	showProvisioningTitle(cmd)

	// Create provisioner and provision cluster
	err = provisionCluster(cmd, cluster, provisioner)
	if err != nil {
		return err
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

// showProvisioningTitle displays the provisioning stage title.
func showProvisioningTitle(cmd *cobra.Command) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Create cluster...",
		Emoji:   "🚀",
		Writer:  cmd.OutOrStdout(),
	})
}

// provisionCluster creates the provisioner and provisions the cluster.
func provisionCluster(
	cmd *cobra.Command,
	cluster *v1alpha1.Cluster,
	provisioner provisionerFactory,
) error {
	distribution := cluster.Spec.Distribution
	distributionConfigPath := cluster.Spec.DistributionConfig
	kubeconfigPath := cluster.Spec.Connection.Kubeconfig

	// Create provisioner based on distribution
	var clusterProvisioner clusterprovisioner.ClusterProvisioner

	var clusterName string

	var err error

	if provisioner != nil {
		clusterProvisioner, clusterName, err = provisioner(
			cmd.Context(),
			distribution,
			distributionConfigPath,
			kubeconfigPath,
		)
	} else {
		// Load config once and get both provisioner and cluster name
		clusterProvisioner, clusterName, err = clusterprovisioner.CreateClusterProvisioner(
			cmd.Context(),
			distribution,
			distributionConfigPath,
			kubeconfigPath,
		)
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

	return nil
}

// provisionerFactory is a function type for creating cluster provisioners (for testing).
// Returns provisioner, cluster name, and error.
type provisionerFactory func(
	context.Context,
	v1alpha1.Distribution,
	string,
	string,
) (clusterprovisioner.ClusterProvisioner, string, error)
