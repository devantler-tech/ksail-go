// Package registry provides registry mirror management for cluster provisioners.
package registry

import (
	"context"
	"fmt"
	"strings"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/client"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Manager handles registry mirror lifecycle for clusters.
type Manager struct {
	registryManager *dockerclient.RegistryManager
	dockerClient    client.APIClient
}

// NewManager creates a new registry Manager.
func NewManager(dockerClient client.APIClient) (*Manager, error) {
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return nil, fmt.Errorf("failed to create registry manager: %w", err)
	}

	return &Manager{
		registryManager: registryMgr,
		dockerClient:    dockerClient,
	}, nil
}

// SetupRegistriesForKind creates mirror registries based on Kind cluster configuration.
func (m *Manager) SetupRegistriesForKind(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	clusterName string,
) error {
	if kindConfig == nil {
		return nil
	}

	registries := extractRegistriesFromKind(kindConfig)
	if len(registries) == 0 {
		return nil
	}

	for _, reg := range registries {
		config := dockerclient.RegistryConfig{
			Name:         reg.Name,
			Port:         reg.Port,
			UpstreamURL:  reg.Upstream,
			ClusterName:  clusterName,
			NetworkName:  "kind", // Kind uses "kind" network
		}

		if err := m.registryManager.CreateRegistry(ctx, config); err != nil {
			return fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
		}
	}

	return nil
}

// SetupRegistriesForK3d creates mirror registries based on K3d cluster configuration.
func (m *Manager) SetupRegistriesForK3d(
	ctx context.Context,
	k3dConfig *k3dv1alpha5.SimpleConfig,
	clusterName string,
) error {
	if k3dConfig == nil {
		return nil
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

		if err := m.registryManager.CreateRegistry(ctx, config); err != nil {
			return fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
		}
	}

	return nil
}

// CleanupRegistries removes registries that are no longer in use.
func (m *Manager) CleanupRegistries(
	ctx context.Context,
	clusterName string,
	registryNames []string,
	deleteVolumes bool,
) error {
	for _, name := range registryNames {
		if err := m.registryManager.DeleteRegistry(ctx, name, clusterName, deleteVolumes); err != nil {
			// Log error but don't fail the entire cleanup
			fmt.Printf("Warning: failed to cleanup registry %s: %v\n", name, err)
		}
	}

	return nil
}

// RegistryInfo holds information about a registry to be created.
type RegistryInfo struct {
	Name     string
	Upstream string
	Port     int
}

// extractRegistriesFromKind extracts registry information from Kind configuration.
func extractRegistriesFromKind(kindConfig *v1alpha4.Cluster) []RegistryInfo {
	var registries []RegistryInfo

	// Kind uses containerdConfigPatches to configure registry mirrors
	for _, patch := range kindConfig.ContainerdConfigPatches {
		// Parse containerd config to extract registry mirrors
		// Format example:
		// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
		//   endpoint = ["http://localhost:5000"]
		
		mirrors := parseContainerdConfig(patch)
		for host, upstream := range mirrors {
			// Generate a simple name from the host
			name := strings.ReplaceAll(host, ".", "-")
			
			registries = append(registries, RegistryInfo{
				Name:     name,
				Upstream: upstream,
				Port:     5000 + len(registries), // Auto-assign ports starting from 5000
			})
		}
	}

	return registries
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

// parseContainerdConfig parses containerd configuration patches to extract registry mirrors.
func parseContainerdConfig(patch string) map[string]string {
	mirrors := make(map[string]string)

	// Simple parser for containerd config format
	// Look for patterns like:
	// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."docker.io"]
	//   endpoint = ["http://registry-1.docker.io"]

	lines := strings.Split(patch, "\n")
	var currentHost string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Match registry mirror section header
		if strings.Contains(line, `registry.mirrors."`) {
			start := strings.Index(line, `mirrors."`) + len(`mirrors."`)
			end := strings.Index(line[start:], `"`)
			if end > 0 {
				currentHost = line[start : start+end]
			}
		}

		// Match endpoint configuration
		if currentHost != "" && strings.Contains(line, "endpoint") {
			// Extract URL from endpoint = ["http://..."]
			if strings.Contains(line, `["`) && strings.Contains(line, `"]`) {
				start := strings.Index(line, `["`) + 2
				end := strings.Index(line[start:], `"`)
				if end > 0 {
					endpoint := line[start : start+end]
					mirrors[currentHost] = endpoint
					currentHost = ""
				}
			}
		}
	}

	return mirrors
}

// SetupRegistries sets up registries based on the cluster configuration.
func (m *Manager) SetupRegistries(
	ctx context.Context,
	clusterCfg *v1alpha1.Cluster,
	kindConfig *v1alpha4.Cluster,
	k3dConfig *k3dv1alpha5.SimpleConfig,
	clusterName string,
) error {
	// Check if mirror registries are enabled
	if !clusterCfg.Spec.IsMirrorRegistriesEnabled() {
		return nil
	}

	switch clusterCfg.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return m.SetupRegistriesForKind(ctx, kindConfig, clusterName)
	case v1alpha1.DistributionK3d:
		return m.SetupRegistriesForK3d(ctx, k3dConfig, clusterName)
	default:
		return nil
	}
}
