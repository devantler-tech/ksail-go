package registry

import (
	"fmt"
	"regexp"
	"strings"

	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ExtractRegistriesFromK3d extracts registry configurations from K3d config.
// It looks at the Registries.Use field which contains registry references.
func ExtractRegistriesFromK3d(cfg *k3dv1alpha5.SimpleConfig) ([]RegistryConfig, error) {
	if cfg == nil || cfg.Registries.Use == nil || len(cfg.Registries.Use) == 0 {
		return []RegistryConfig{}, nil
	}

	var registries []RegistryConfig

	for _, registryRef := range cfg.Registries.Use {
		reg, err := parseK3dRegistryReference(registryRef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse registry reference '%s': %w", registryRef, err)
		}
		registries = append(registries, reg)
	}

	return registries, nil
}

// parseK3dRegistryReference parses a K3d registry reference.
// Format: "k3d-<name>:<port>" or "<name>:<port>"
func parseK3dRegistryReference(ref string) (RegistryConfig, error) {
	parts := strings.Split(ref, ":")
	if len(parts) != 2 {
		return RegistryConfig{}, fmt.Errorf("invalid registry reference format: %s", ref)
	}

	name := parts[0]
	port := parts[1]

	return RegistryConfig{
		Name:     name,
		HostPort: port,
		Image:    DefaultRegistryImage,
	}, nil
}

// ExtractRegistriesFromKind extracts registry configurations from Kind containerd patches.
// It parses ContainerdConfigPatches looking for registry mirror configurations.
func ExtractRegistriesFromKind(cfg *v1alpha4.Cluster) ([]RegistryConfig, error) {
	if cfg == nil || cfg.ContainerdConfigPatches == nil || len(cfg.ContainerdConfigPatches) == 0 {
		return []RegistryConfig{}, nil
	}

	var registries []RegistryConfig
	seen := make(map[string]bool)

	for _, patch := range cfg.ContainerdConfigPatches {
		regs, err := parseContainerdPatch(patch)
		if err != nil {
			return nil, fmt.Errorf("failed to parse containerd patch: %w", err)
		}

		// Deduplicate registries
		for _, reg := range regs {
			key := reg.Name + ":" + reg.HostPort
			if !seen[key] {
				seen[key] = true
				registries = append(registries, reg)
			}
		}
	}

	return registries, nil
}

// parseContainerdPatch parses a containerd config patch for registry mirrors.
// Looks for patterns like:
//
//	[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
//	  endpoint = ["http://registry:5000"]
func parseContainerdPatch(patch string) ([]RegistryConfig, error) {
	var registries []RegistryConfig

	// Pattern to match registry mirror configuration
	// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."<host>:<port>"]
	mirrorPattern := regexp.MustCompile(`\[plugins\."io\.containerd\.grpc\.v1\.cri"\.registry\.mirrors\."([^"]+)"\]`)
	
	// Pattern to match endpoint = ["http://<name>:<port>"] or ["http://<name>"]
	endpointPattern := regexp.MustCompile(`endpoint\s*=\s*\[\s*"https?://([^:"]+)(?::(\d+))?"`)

	// Find all mirror sections
	mirrorMatches := mirrorPattern.FindAllStringSubmatch(patch, -1)
	endpointMatches := endpointPattern.FindAllStringSubmatch(patch, -1)

	if len(mirrorMatches) == 0 || len(endpointMatches) == 0 {
		return registries, nil
	}

	// For each mirror, try to find corresponding endpoint
	for i, mirrorMatch := range mirrorMatches {
		if len(mirrorMatch) < 2 {
			continue
		}

		mirrorHost := mirrorMatch[1] // e.g., "localhost:5000"

		// Try to find endpoint for this mirror
		if i < len(endpointMatches) && len(endpointMatches[i]) >= 2 {
			endpointName := endpointMatches[i][1] // e.g., "registry" or "k3d-registry"

			// Extract port from mirror host if present
			mirrorParts := strings.Split(mirrorHost, ":")
			hostPort := "5000" // default
			if len(mirrorParts) == 2 {
				hostPort = mirrorParts[1]
			}

			registries = append(registries, RegistryConfig{
				Name:     endpointName,
				HostPort: hostPort,
				Image:    DefaultRegistryImage,
			})
		}
	}

	return registries, nil
}
