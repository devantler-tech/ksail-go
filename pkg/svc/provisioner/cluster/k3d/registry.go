package k3dprovisioner

import (
	"context"
	"fmt"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// RegistryInfo holds information about a registry to be created.
type RegistryInfo struct {
	Name     string
	Upstream string
	Port     int
}

// SetupRegistries creates mirror registries based on K3d cluster configuration.
func SetupRegistries(
	ctx context.Context,
	k3dConfig *k3dv1alpha5.SimpleConfig,
	clusterName string,
	dockerClient client.APIClient,
) error {
	if k3dConfig == nil {
		return nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	registries := extractRegistriesFromK3d(k3dConfig)
	if len(registries) == 0 {
		return nil
	}

	for _, reg := range registries {
		config := dockerclient.RegistryConfig{
			Name:         reg.Name,
			Port:         reg.Port,
			UpstreamURL:  reg.Upstream,
			ClusterName:  clusterName,
			NetworkName:  fmt.Sprintf("k3d-%s", clusterName), // K3d uses "k3d-{clustername}" network
		}

		if err := registryMgr.CreateRegistry(ctx, config); err != nil {
			return fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
		}
	}

	return nil
}

// CleanupRegistries removes registries that are no longer in use.
func CleanupRegistries(
	ctx context.Context,
	k3dConfig *k3dv1alpha5.SimpleConfig,
	clusterName string,
	dockerClient client.APIClient,
	deleteVolumes bool,
) error {
	if k3dConfig == nil {
		return nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	registries := extractRegistriesFromK3d(k3dConfig)
	if len(registries) == 0 {
		return nil
	}

	for _, reg := range registries {
		if err := registryMgr.DeleteRegistry(ctx, reg.Name, clusterName, deleteVolumes); err != nil {
			// Log error but don't fail the entire cleanup
			fmt.Printf("Warning: failed to cleanup registry %s: %v\n", reg.Name, err)
		}
	}

	return nil
}

// extractRegistriesFromK3d extracts registry information from K3d configuration.
func extractRegistriesFromK3d(k3dConfig *k3dv1alpha5.SimpleConfig) []RegistryInfo {
	var registries []RegistryInfo

	// K3d has native registry mirror support
	if k3dConfig.Registries.Use != nil {
		for _, regName := range k3dConfig.Registries.Use {
			// K3d registry names are typically in format: k3d-myregistry:5000
			// We extract the relevant information
			registries = append(registries, RegistryInfo{
				Name:     regName,
				Upstream: "", // K3d manages its own registries
				Port:     5000 + len(registries),
			})
		}
	}

	// Check for registry config
	if k3dConfig.Registries.Config != "" {
		// Registry config file specifies mirrors
		// This would need to be parsed, but for now we'll support the basic case
	}

	return registries
}
