package cluster

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewStartCmd creates and returns the start command.
func NewStartCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"start",
		"Start a stopped cluster",
		`Start a previously stopped cluster.`,
		HandleStartRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}

// HandleStartRunE handles the start command.
func HandleStartRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	return handleStartRunEWithProvisioner(cmd, manager, nil)
}

// handleStartRunEWithProvisioner is the internal implementation that accepts an optional provisioner for testing.
func handleStartRunEWithProvisioner(
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
		return fmt.Errorf("failed to load cluster configuration: %w", err)
	}

	// Start the cluster
	err = startCluster(cmd, cluster, provisioner, tmr)
	if err != nil {
		return err
	}

	// Display success with timing
	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster started",
		Timer:      tmr,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// startCluster creates the provisioner and starts the cluster.
func startCluster(
	cmd *cobra.Command,
	cluster *v1alpha1.Cluster,
	provisioner provisionerFactory,
	tmr timer.Timer,
) error {
	// Transition to starting stage
	tmr.NewStage()

	// Show starting title
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Start cluster...",
		Emoji:   "▶️",
		Writer:  cmd.OutOrStdout(),
	})

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
		Content: "starting cluster",
		Writer:  cmd.OutOrStdout(),
	})

	// Start the cluster
	err = clusterProvisioner.Start(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to start cluster: %w", err)
	}

	return nil
}
