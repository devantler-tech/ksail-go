package registry

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var (
	// ErrInvalidRegistryFormat is returned when a registry reference has an invalid format.
	ErrInvalidRegistryFormat = errors.New("invalid registry reference format")
)

// ExtractRegistriesFromK3d extracts registry configurations from K3d config.
// It looks at the Registries.Use field which contains registry references.
func ExtractRegistriesFromK3d(cfg *k3dv1alpha5.SimpleConfig) ([]Config, error) {
	if cfg == nil || cfg.Registries.Use == nil || len(cfg.Registries.Use) == 0 {
		return []Config{}, nil
	}

	registries := make([]Config, 0, len(cfg.Registries.Use))

	for _, registryRef := range cfg.Registries.Use {
		reg, err := parseK3dRegistryReference(registryRef)
		if err != nil {
			return nil, fmt.Errorf("failed to parse registry reference '%s': %w", registryRef, err)
		}

		registries = append(registries, reg)
	}

	return registries, nil
}

const registryReferencePartsCount = 2

// parseK3dRegistryReference parses a K3d registry reference.
// Format: "k3d-<name>:<port>" or "<name>:<port>".
func parseK3dRegistryReference(ref string) (Config, error) {
	parts := strings.Split(ref, ":")
	if len(parts) != registryReferencePartsCount {
		return Config{}, fmt.Errorf("%w: %s", ErrInvalidRegistryFormat, ref)
	}

	name := parts[0]
	port := parts[1]

	return Config{
		Name:     name,
		HostPort: port,
		Image:    DefaultRegistryImage,
	}, nil
}

// ExtractRegistriesFromKind extracts registry configurations from Kind containerd patches.
// It parses ContainerdConfigPatches looking for registry mirror configurations.
func ExtractRegistriesFromKind(cfg *v1alpha4.Cluster) ([]Config, error) {
	if cfg == nil || cfg.ContainerdConfigPatches == nil || len(cfg.ContainerdConfigPatches) == 0 {
		return []Config{}, nil
	}

	var registries []Config

	seen := make(map[string]bool)

	for _, patch := range cfg.ContainerdConfigPatches {
		regs := parseContainerdPatch(patch)

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

const (
	defaultRegistryPort      = "5000"
	mirrorHostPartsWithPort  = 2
	minEndpointMatchParts    = 2
	minExpectedPartsInMirror = 2
)

// parseContainerdPatch parses a containerd config patch for registry mirrors.
// Looks for patterns like:
//
//	[plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:5000"]
//	  endpoint = ["http://registry:5000"]
func parseContainerdPatch(patch string) []Config {
	var registries []Config

	// Pattern to match registry mirror configuration
	// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."<host>:<port>"]
	mirrorPattern := regexp.MustCompile(`\[plugins\."io\.containerd\.grpc\.v1\.cri"\.registry\.mirrors\."([^"]+)"\]`)

	// Pattern to match endpoint = ["http://<name>:<port>"] or ["http://<name>"]
	endpointPattern := regexp.MustCompile(`endpoint\s*=\s*\[\s*"https?://([^:"]+)(?::(\d+))?"`)

	// Find all mirror sections
	mirrorMatches := mirrorPattern.FindAllStringSubmatch(patch, -1)
	endpointMatches := endpointPattern.FindAllStringSubmatch(patch, -1)

	if len(mirrorMatches) == 0 || len(endpointMatches) == 0 {
		return registries
	}

	// For each mirror, try to find corresponding endpoint
	for index, mirrorMatch := range mirrorMatches {
		if len(mirrorMatch) < minExpectedPartsInMirror {
			continue
		}

		mirrorHost := mirrorMatch[1] // e.g., "localhost:5000"

		// Try to find endpoint for this mirror
		if index < len(endpointMatches) && len(endpointMatches[index]) >= minEndpointMatchParts {
			endpointName := endpointMatches[index][1] // e.g., "registry" or "k3d-registry"

			// Extract port from mirror host if present
			mirrorParts := strings.Split(mirrorHost, ":")

			hostPort := defaultRegistryPort // default
			if len(mirrorParts) == mirrorHostPartsWithPort {
				hostPort = mirrorParts[1]
			}

			registries = append(registries, Config{
				Name:     endpointName,
				HostPort: hostPort,
				Image:    DefaultRegistryImage,
			})
		}
	}

	return registries
}
