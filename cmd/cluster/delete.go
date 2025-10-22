package cluster

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
)

// newDeleteLifecycleConfig creates the lifecycle configuration for cluster deletion.
func newDeleteLifecycleConfig() shared.LifecycleConfig {
	return shared.LifecycleConfig{
		TitleEmoji:         "🗑️",
		TitleContent:       "Delete cluster...",
		ActivityContent:    "deleting cluster",
		SuccessContent:     "cluster deleted",
		ErrorMessagePrefix: "failed to delete cluster",
		Action: func(ctx context.Context, provisioner clusterprovisioner.ClusterProvisioner, clusterName string) error {
			return provisioner.Delete(ctx, clusterName)
		},
	}
}

// NewDeleteCmd creates and returns the delete command.
func NewDeleteCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Destroy a cluster",
		Long:         `Destroy a cluster.`,
		SilenceUsage: true,
	}

	cfgManager := ksailconfigmanager.NewCommandConfigManager(
		cmd,
		ksailconfigmanager.DefaultClusterFieldSelectors(),
	)

	cmd.RunE = newDeleteCommandRunE(runtimeContainer, cfgManager)

	return cmd
}

// newDeleteCommandRunE creates the RunE handler for cluster deletion with registry cleanup.
func newDeleteCommandRunE(
	runtimeContainer *runtime.Runtime,
	cfgManager *ksailconfigmanager.ConfigManager,
) func(*cobra.Command, []string) error {
	return shared.WrapLifecycleHandler(runtimeContainer, cfgManager, handleDeleteRunE)
}

// handleDeleteRunE executes cluster deletion with registry cleanup.
func handleDeleteRunE(
	cmd *cobra.Command,
	cfgManager *ksailconfigmanager.ConfigManager,
	deps shared.LifecycleDeps,
) error {
	config := newDeleteLifecycleConfig()

	// Execute cluster deletion
	err := shared.HandleLifecycleRunE(cmd, cfgManager, deps, config)
	if err != nil {
		return fmt.Errorf("cluster deletion failed: %w", err)
	}

	// Clean up registries after cluster deletion
	clusterCfg := cfgManager.GetConfig()

	err = cleanupMirrorRegistries(cmd, clusterCfg, deps)
	if err != nil {
		// Log warning but don't fail the delete operation
		notify.WriteMessage(notify.Message{
			Type:    notify.WarningType,
			Content: fmt.Sprintf("failed to cleanup registries: %v", err),
			Writer:  cmd.OutOrStdout(),
		})
	}

	return nil
}

// cleanupMirrorRegistries cleans up registries for Kind after cluster deletion.
// K3d handles registry cleanup natively through its own configuration.
func cleanupMirrorRegistries(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
) error {
	// Only Kind requires registry cleanup - K3d handles it natively
	if clusterCfg.Spec.Distribution != v1alpha1.DistributionKind {
		return nil
	}

	// Load Kind config to check if containerd patches exist
	kindConfigMgr := kindconfigmanager.NewConfigManager(clusterCfg.Spec.DistributionConfig)

	err := kindConfigMgr.LoadConfig(deps.Timer)
	if err != nil {
		return fmt.Errorf("failed to load kind config: %w", err)
	}

	kindConfig := kindConfigMgr.GetConfig()

	// If no containerd patches, no registries to clean up
	if len(kindConfig.ContainerdConfigPatches) == 0 {
		return nil
	}

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	defer func() {
		closeErr := dockerClient.Close()
		if closeErr != nil {
			// Log error but don't fail the operation
			notify.WriteMessage(notify.Message{
				Type:    notify.WarningType,
				Content: fmt.Sprintf("failed to close docker client: %v", closeErr),
				Writer:  cmd.OutOrStdout(),
			})
		}
	}()

	// Display activity message
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "cleaning up mirror registries",
		Writer:  cmd.OutOrStdout(),
	})

	// Always delete volumes (can be made configurable later)
	deleteVolumes := false

	// Clean up registries for Kind (cluster name comes from kindConfig.Name)
	err = kindprovisioner.CleanupRegistries(cmd.Context(), kindConfig, kindConfig.Name, dockerClient, deleteVolumes)
	if err != nil {
		return fmt.Errorf("failed to cleanup registries: %w", err)
	}

	return nil
}
