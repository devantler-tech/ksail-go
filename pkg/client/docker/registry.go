package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

const (
	// DefaultRegistryImage is the default Docker registry image to use.
	DefaultRegistryImage = "registry:3"
)

// Error definitions for registry operations.
var (
	// ErrInvalidRegistryFormat is returned when a registry reference has an invalid format.
	ErrInvalidRegistryFormat = errors.New("invalid registry reference format")
)

// RegistryConfig describes a registry container configuration.
type RegistryConfig struct {
	Name     string // Container name
	HostPort string // Host port to bind (e.g., "5000")
	Image    string // Registry image (default: registry:3)
}

// RegistryManager handles Docker registry container lifecycle.
type RegistryManager struct {
	client client.APIClient
}

// NewRegistryManager creates a new registry manager with the provided Docker client.
func NewRegistryManager(dockerClient client.APIClient) *RegistryManager {
	return &RegistryManager{
		client: dockerClient,
	}
}

// CreateRegistry creates a Docker registry container if it doesn't already exist.
// Returns nil if the registry already exists or was successfully created.
func (m *RegistryManager) CreateRegistry(ctx context.Context, cfg RegistryConfig) error {
	if cfg.Image == "" {
		cfg.Image = DefaultRegistryImage
	}

	// Check if container already exists
	exists, err := m.containerExists(ctx, cfg.Name)
	if err != nil {
		return fmt.Errorf("failed to check if registry exists: %w", err)
	}

	if exists {
		return nil // Registry already exists
	}

	// Pull the registry image
	err = m.pullImage(ctx, cfg.Image)
	if err != nil {
		return fmt.Errorf("failed to pull registry image: %w", err)
	}

	// Create container configuration
	containerPort := nat.Port("5000/tcp")
	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: cfg.HostPort,
	}

	containerConfig := &container.Config{
		Image: cfg.Image,
		ExposedPorts: nat.PortSet{
			containerPort: struct{}{},
		},
	}

	hostConfig := &container.HostConfig{
		PortBindings: nat.PortMap{
			containerPort: []nat.PortBinding{hostBinding},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
	}

	networkConfig := &network.NetworkingConfig{}

	// Create the container
	resp, err := m.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		cfg.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to create registry container: %w", err)
	}

	// Start the container
	err = m.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}

	return nil
}

// containerExists checks if a container with the given name exists.
func (m *RegistryManager) containerExists(ctx context.Context, name string) (bool, error) {
	containers, err := m.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return false, fmt.Errorf("failed to list containers: %w", err)
	}

	for _, c := range containers {
		for _, n := range c.Names {
			// Docker container names start with "/"
			cleanName := strings.TrimPrefix(n, "/")
			if cleanName == name {
				return true, nil
			}
		}
	}

	return false, nil
}

// pullImage pulls a Docker image if not already present.
func (m *RegistryManager) pullImage(ctx context.Context, imageName string) error {
	reader, err := m.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}

	defer func() {
		_ = reader.Close()
	}()

	// Consume the reader to ensure the pull completes
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to read pull response: %w", err)
	}

	return nil
}

// ExtractRegistriesFromK3d extracts registry configurations from K3d config.
// It looks at the Registries.Use field which contains registry references.
func ExtractRegistriesFromK3d(cfg *k3dv1alpha5.SimpleConfig) ([]RegistryConfig, error) {
	if cfg == nil || cfg.Registries.Use == nil || len(cfg.Registries.Use) == 0 {
		return []RegistryConfig{}, nil
	}

	registries := make([]RegistryConfig, 0, len(cfg.Registries.Use))

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
func parseK3dRegistryReference(ref string) (RegistryConfig, error) {
	parts := strings.Split(ref, ":")
	if len(parts) != registryReferencePartsCount {
		return RegistryConfig{}, fmt.Errorf("%w: %s", ErrInvalidRegistryFormat, ref)
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
func parseContainerdPatch(patch string) []RegistryConfig {
	var registries []RegistryConfig

	// Pattern to match registry mirror configuration
	// [plugins."io.containerd.grpc.v1.cri".registry.mirrors."<host>:<port>"]
	mirrorPattern := regexp.MustCompile(
		`\[plugins\."io\.containerd\.grpc\.v1\.cri"\.registry\.mirrors\."([^"]+)"\]`,
	)

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

			registries = append(registries, RegistryConfig{
				Name:     endpointName,
				HostPort: hostPort,
				Image:    DefaultRegistryImage,
			})
		}
	}

	return registries
}
