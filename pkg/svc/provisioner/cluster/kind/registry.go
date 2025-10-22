// Package kindprovisioner provides implementations of the Provisioner interface
// for provisioning clusters in different providers.
package kindprovisioner

import (
	"context"
	"fmt"
	"strings"

	dockerclient "github.com/devantler-tech/ksail-go/pkg/client/docker"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// RegistryInfo holds information about a registry to be created.
type RegistryInfo struct {
	Name     string
	Upstream string
	Port     int
}

// SetupRegistries creates mirror registries based on Kind cluster configuration.
func SetupRegistries(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	clusterName string,
	dockerClient client.APIClient,
) error {
	if kindConfig == nil {
		return nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
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

		if err := registryMgr.CreateRegistry(ctx, config); err != nil {
			return fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
		}
	}

	return nil
}

// CleanupRegistries removes registries that are no longer in use.
func CleanupRegistries(
	ctx context.Context,
	kindConfig *v1alpha4.Cluster,
	clusterName string,
	dockerClient client.APIClient,
	deleteVolumes bool,
) error {
	if kindConfig == nil {
		return nil
	}

	// Create registry manager
	registryMgr, err := dockerclient.NewRegistryManager(dockerClient)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	registries := extractRegistriesFromKind(kindConfig)
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
