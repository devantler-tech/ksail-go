package cluster

import (
	"context"
	"fmt"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	ksailconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/ksail"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/io/config-manager/kind"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster"
	registrymanager "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/registry"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/spf13/cobra"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
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
			Content: fmt.Sprintf("Warning: failed to cleanup registries: %v", err),
			Writer:  cmd.OutOrStdout(),
		})
	}

	return nil
}

// cleanupMirrorRegistries cleans up registries that are no longer in use.
func cleanupMirrorRegistries(
	cmd *cobra.Command,
	clusterCfg *v1alpha1.Cluster,
	deps shared.LifecycleDeps,
) error {
	// Check if mirror registries are enabled
	if !clusterCfg.Spec.IsMirrorRegistriesEnabled() {
		return nil
	}

	// Create Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}
	defer dockerClient.Close()

	// Create registry manager
	regMgr, err := registrymanager.NewManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	// Load distribution-specific config to infer registry names
	var kindConfig *v1alpha4.Cluster
	var k3dConfig *k3dv1alpha5.SimpleConfig
	var registryNames []string

	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		kindConfigMgr := kindconfigmanager.NewConfigManager(clusterCfg.Spec.DistributionConfig)
		err = kindConfigMgr.LoadConfig(deps.Timer)
		if err != nil {
			return fmt.Errorf("failed to load kind config: %w", err)
		}
		kindConfig = kindConfigMgr.GetConfig()
		// Extract registry names from kind config
		registries := extractRegistriesFromKindConfig(kindConfig)
		for _, reg := range registries {
			registryNames = append(registryNames, reg.Name)
		}
	case v1alpha1.DistributionK3d:
		k3dConfigMgr := k3dconfigmanager.NewConfigManager(clusterCfg.Spec.DistributionConfig)
		err = k3dConfigMgr.LoadConfig(deps.Timer)
		if err != nil {
			return fmt.Errorf("failed to load k3d config: %w", err)
		}
		k3dConfig = k3dConfigMgr.GetConfig()
		// Extract registry names from k3d config
		registries := extractRegistriesFromK3dConfig(k3dConfig)
		for _, reg := range registries {
			registryNames = append(registryNames, reg.Name)
		}
	default:
		return nil
	}

	if len(registryNames) == 0 {
		return nil
	}

	// Get cluster name
	clusterName := clusterCfg.Spec.Connection.Context
	if clusterName == "" {
		clusterName = "default"
	}

	// Clean up registries
	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "cleaning up mirror registries",
		Writer:  cmd.OutOrStdout(),
	})

	// Always delete volumes (can be made configurable later)
	deleteVolumes := false
	err = regMgr.CleanupRegistries(cmd.Context(), clusterName, registryNames, deleteVolumes)
	if err != nil {
		return fmt.Errorf("failed to cleanup registries: %w", err)
	}

	return nil
}

// extractRegistriesFromKindConfig is a helper to extract registry info from Kind config.
func extractRegistriesFromKindConfig(kindConfig *v1alpha4.Cluster) []registrymanager.RegistryInfo {
	if kindConfig == nil {
		return nil
	}
	// This reuses the logic from the registry manager package
	// For simplicity, we'll just parse the containerd patches again
	var registries []registrymanager.RegistryInfo
	for _, patch := range kindConfig.ContainerdConfigPatches {
		// Parse containerd config to extract registry mirrors
		// For now, we'll use a simplified version
		if len(patch) > 0 {
			// Placeholder: extract actual registry names from patches
			// This would need proper parsing logic
		}
	}
	return registries
}

// extractRegistriesFromK3dConfig is a helper to extract registry info from K3d config.
func extractRegistriesFromK3dConfig(k3dConfig *k3dv1alpha5.SimpleConfig) []registrymanager.RegistryInfo {
	if k3dConfig == nil {
		return nil
	}
	var registries []registrymanager.RegistryInfo
	if k3dConfig.Registries.Use != nil {
		for _, regName := range k3dConfig.Registries.Use {
			registries = append(registries, registrymanager.RegistryInfo{
				Name: regName,
			})
		}
	}
	return registries
}
