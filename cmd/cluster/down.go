package cluster

import (
	"fmt"

	helpers "github.com/devantler-tech/ksail-go/cmd/internal/helpers"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	configmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/ksail"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/ui/timer"
	"github.com/spf13/cobra"
)

// NewDownCmd creates and returns the down command.
func NewDownCmd() *cobra.Command {
	return helpers.NewCobraCommand(
		"down",
		"Destroy a cluster",
		`Destroy a cluster.`,
		HandleDownRunE,
		configmanager.StandardDistributionFieldSelector(),
		configmanager.StandardContextFieldSelector(),
	)
}

// HandleDownRunE handles the down command.
// Exported for testing purposes.
func HandleDownRunE(
	cmd *cobra.Command,
	manager *configmanager.ConfigManager,
	_ []string,
) error {
	return handleDownRunEWithProvisioner(cmd, manager, nil)
}

// handleDownRunEWithProvisioner is the internal implementation that accepts an optional provisioner for testing.
func handleDownRunEWithProvisioner(
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

	// Transition to deletion stage
	tmr.NewStage()

	// Delete the cluster
	err = deleteCluster(cmd, cluster, provisioner)
	if err != nil {
		return err
	}

	// Display success with timing
	notify.WriteMessage(notify.Message{
		Type:       notify.SuccessType,
		Content:    "cluster deleted",
		Timer:      tmr,
		Writer:     cmd.OutOrStdout(),
		MultiStage: true,
	})

	return nil
}

// showDeletionTitle displays the deletion stage title.
func showDeletionTitle(cmd *cobra.Command) {
	cmd.Println()
	notify.WriteMessage(notify.Message{
		Type:    notify.TitleType,
		Content: "Delete cluster...",
		Emoji:   "🗑️",
		Writer:  cmd.OutOrStdout(),
	})
}

// deleteCluster creates the provisioner and deletes the cluster.
func deleteCluster(
	cmd *cobra.Command,
	cluster *v1alpha1.Cluster,
	provisioner provisionerFactory,
) error {
	// Show deletion title
	showDeletionTitle(cmd)

	// Create provisioner
	clusterProvisioner, clusterName, err := createProvisionerForCluster(
		cmd.Context(),
		cluster,
		provisioner,
	)
	if err != nil {
		return err
	}

	// Show activity message
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "deleting cluster",
		Writer:  cmd.OutOrStdout(),
	})

	// Delete the cluster
	err = clusterProvisioner.Delete(cmd.Context(), clusterName)
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %w", err)
	}

	return nil
}
