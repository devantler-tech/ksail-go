package cluster

import (
	"context"
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewStopCmd creates and returns the stop command.
func NewStopCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"stop",
		"Stop the Kubernetes cluster",
		`Stop the Kubernetes cluster without removing it.`,
		HandleStopRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}

// HandleStopRunE handles the stop command.
func HandleStopRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	return handleStopRunEWithProvisioner(cmd, manager, nil)
}

// handleStopRunEWithProvisioner is the internal implementation that accepts an optional provisioner for testing.
func handleStopRunEWithProvisioner(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	provisioner stopProvisionerFactory,
) error {
	// Start timing
	tmr := timer.New()
	tmr.Start()

	// Load cluster configuration
	cluster, err := manager.LoadConfig(tmr)
	if err != nil {
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Stop the cluster
	err = stopCluster(cmd, cluster, provisioner)
	if err != nil {
		return err
	}

	// Display success with timing
	notify.WriteMessage(notify.Message{
		Type:    notify.SuccessType,
		Content: "cluster stopped",
		Timer:   tmr,
		Writer:  cmd.OutOrStdout(),
	})

	return nil
}

// stopCluster creates the provisioner and stops the cluster.
func stopCluster(
	cmd *cobra.Command,
	cluster *v1alpha1.Cluster,
	provisioner stopProvisionerFactory,
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

	// Stop the cluster
	err = clusterProvisioner.Stop(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to stop cluster: %w", err)
	}

	return nil
}

// stopProvisionerFactory is a function type for creating cluster provisioners (for testing).
// Returns provisioner, cluster name, and error.
type stopProvisionerFactory func(
	context.Context,
	v1alpha1.Distribution,
	string,
	string,
) (clusterprovisioner.ClusterProvisioner, string, error)
