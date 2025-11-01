package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

// Registry error definitions.
var (
	// ErrRegistryNotFound is returned when a registry container is not found.
	ErrRegistryNotFound = errors.New("registry not found")
	// ErrRegistryAlreadyExists is returned when trying to create a registry that already exists.
	ErrRegistryAlreadyExists = errors.New("registry already exists")
	// ErrRegistryPortNotFound is returned when the registry port cannot be determined.
	ErrRegistryPortNotFound = errors.New("registry port not found")
)

const (
	// RegistryImageName is the default registry image to use.
	RegistryImageName = "registry:3"
	// RegistryLabelKey marks registry containers as managed by ksail.
	RegistryLabelKey = "io.ksail.registry"
	// DefaultRegistryPort is the default port for registry containers.
	DefaultRegistryPort = 5000
	// RegistryPortBase is the base port number for calculating registry ports.
	RegistryPortBase = 5000
	// HostPortParts is the expected number of parts in a host:port string.
	HostPortParts = 2
)

// RegistryManager manages Docker registry containers for mirror/pull-through caching.
type RegistryManager struct {
	client client.APIClient
}

// NewRegistryManager creates a new RegistryManager.
func NewRegistryManager(apiClient client.APIClient) (*RegistryManager, error) {
	if apiClient == nil {
		return nil, ErrAPIClientNil
	}

	return &RegistryManager{
		client: apiClient,
	}, nil
}

// RegistryConfig holds configuration for creating a registry.
type RegistryConfig struct {
	Name        string
	Port        int
	UpstreamURL string
	ClusterName string
	NetworkName string
	VolumeName  string
}

// CreateRegistry creates a registry container with the given configuration.
// If the registry already exists, it returns ErrRegistryAlreadyExists.
func (rm *RegistryManager) CreateRegistry(ctx context.Context, config RegistryConfig) error {
	// Check if registry already exists
	exists, err := rm.registryExists(ctx, config.Name)
	if err != nil {
		return fmt.Errorf("failed to check if registry exists: %w", err)
	}

	if exists {
		// Add cluster label to existing registry
		return rm.addClusterLabel(ctx, config.Name, config.ClusterName)
	}

	// Pull registry image if not present
	err = rm.ensureRegistryImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure registry image: %w", err)
	}

	// Create volume for registry data using a distribution-agnostic name for reuse
	volumeName := rm.resolveVolumeName(config)
	if volumeName == "" {
		volumeName = config.Name
	}

	err = rm.createVolume(ctx, volumeName)
	if err != nil {
		return fmt.Errorf("failed to create registry volume: %w", err)
	}

	// Prepare container configuration
	containerConfig := rm.buildContainerConfig(config)
	hostConfig := rm.buildHostConfig(config, volumeName)
	networkConfig := rm.buildNetworkConfig(config)

	// Use provided registry name directly so other components can reference it
	containerName := config.Name

	// Create container
	resp, err := rm.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		networkConfig,
		nil,
		containerName,
	)
	if err != nil {
		return fmt.Errorf("failed to create registry container: %w", err)
	}

	// Start container
	err = rm.client.ContainerStart(ctx, resp.ID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}

	return nil
}

// DeleteRegistry removes a registry container and optionally its volume.
// If deleteVolume is true, the associated volume will be removed.
// If the registry is still in use by other clusters, it returns an error.
func (rm *RegistryManager) DeleteRegistry(
	ctx context.Context,
	name, _ string,
	deleteVolume bool,
	networkName string,
) error {
	containers, err := rm.listRegistryContainers(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to list registry containers: %w", err)
	}

	if len(containers) == 0 {
		return ErrRegistryNotFound
	}

	registryContainer := containers[0]

	trimmedNetwork := strings.TrimSpace(networkName)

	inspect, err := inspectContainer(ctx, rm.client, registryContainer.ID)
	if err != nil {
		return err
	}

	inspect, err = disconnectRegistryNetwork(
		ctx,
		rm.client,
		registryContainer.ID,
		name,
		trimmedNetwork,
		inspect,
	)
	if err != nil {
		return err
	}

	if registryAttachedToOtherClusters(inspect, trimmedNetwork) {
		return nil
	}

	stopErr := rm.stopRegistryContainer(ctx, registryContainer)
	if stopErr != nil {
		return stopErr
	}

	removeErr := rm.client.ContainerRemove(ctx, registryContainer.ID, container.RemoveOptions{})
	if removeErr != nil {
		return fmt.Errorf("failed to remove registry container: %w", removeErr)
	}

	return cleanupRegistryVolume(ctx, rm.client, registryContainer, name, deleteVolume)
}

// ListRegistries returns a list of all ksail registry containers.
func (rm *RegistryManager) ListRegistries(ctx context.Context) ([]string, error) {
	containers, err := rm.listAllRegistryContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list registry containers: %w", err)
	}

	registries := make([]string, 0, len(containers))

	seen := make(map[string]struct{}, len(containers))
	for _, c := range containers {
		name := c.Labels[RegistryLabelKey]
		if name == "" {
			for _, rawName := range c.Names {
				trimmed := strings.TrimPrefix(rawName, "/")
				if trimmed != "" {
					name = trimmed

					break
				}
			}
		}

		if name == "" {
			continue
		}

		if _, exists := seen[name]; exists {
			continue
		}

		seen[name] = struct{}{}
		registries = append(registries, name)
	}

	return registries, nil
}

// IsRegistryInUse checks if a registry is being used by any clusters.
// A registry is considered in use if it exists and is running.
func (rm *RegistryManager) IsRegistryInUse(ctx context.Context, name string) (bool, error) {
	containers, err := rm.listRegistryContainers(ctx, name)
	if err != nil {
		return false, fmt.Errorf("failed to list registry containers: %w", err)
	}

	if len(containers) == 0 {
		return false, nil
	}

	// Check if container is running
	return containers[0].State == "running", nil
}

// GetRegistryPort returns the host port for a registry.
func (rm *RegistryManager) GetRegistryPort(ctx context.Context, name string) (int, error) {
	containers, err := rm.listRegistryContainers(ctx, name)
	if err != nil {
		return 0, fmt.Errorf("failed to list registry containers: %w", err)
	}

	if len(containers) == 0 {
		return 0, ErrRegistryNotFound
	}

	// Get port from container ports
	for _, port := range containers[0].Ports {
		if port.PrivatePort == DefaultRegistryPort {
			return int(port.PublicPort), nil
		}
	}

	return 0, ErrRegistryPortNotFound
}

// Helper methods

func (rm *RegistryManager) registryExists(ctx context.Context, name string) (bool, error) {
	containers, err := rm.listRegistryContainers(ctx, name)
	if err != nil {
		return false, err
	}

	return len(containers) > 0, nil
}

func (rm *RegistryManager) listRegistryContainers(
	ctx context.Context,
	name string,
) ([]container.Summary, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("name", name)
	filterArgs.Add("ancestor", RegistryImageName)
	filterArgs.Add("label", fmt.Sprintf("%s=%s", RegistryLabelKey, name))

	containers, err := rm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list registry containers: %w", err)
	}

	return containers, nil
}

func (rm *RegistryManager) listAllRegistryContainers(
	ctx context.Context,
) ([]container.Summary, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("ancestor", RegistryImageName)
	filterArgs.Add("label", RegistryLabelKey)

	containers, err := rm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list all registry containers: %w", err)
	}

	return containers, nil
}

func (rm *RegistryManager) ensureRegistryImage(ctx context.Context) error {
	// Check if image exists
	_, err := rm.client.ImageInspect(ctx, RegistryImageName)
	if err == nil {
		return nil
	}

	// Pull image
	reader, err := rm.client.ImagePull(ctx, RegistryImageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull registry image: %w", err)
	}

	// Consume pull output
	_, err = io.Copy(io.Discard, reader)
	closeErr := reader.Close()

	if err != nil {
		return fmt.Errorf("failed to read image pull output: %w", err)
	}

	if closeErr != nil {
		return fmt.Errorf("failed to close image pull reader: %w", closeErr)
	}

	return nil
}

func (rm *RegistryManager) createVolume(
	ctx context.Context,
	volumeName string,
) error {
	// Check if volume already exists
	_, err := rm.client.VolumeInspect(ctx, volumeName)
	if err == nil {
		return nil // Volume already exists
	}

	// Create volume
	_, err = rm.client.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
	})
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	return nil
}

func (rm *RegistryManager) buildContainerConfig(
	config RegistryConfig,
) *container.Config {
	env := []string{}
	if config.UpstreamURL != "" {
		env = append(env, "REGISTRY_PROXY_REMOTEURL="+config.UpstreamURL)
	}

	labels := map[string]string{}
	if config.Name != "" {
		labels[RegistryLabelKey] = config.Name
	}

	return &container.Config{
		Image: RegistryImageName,
		Env:   env,
		ExposedPorts: nat.PortSet{
			"5000/tcp": struct{}{},
		},
		Labels: labels,
	}
}

func (rm *RegistryManager) buildHostConfig(
	config RegistryConfig,
	volumeName string,
) *container.HostConfig {
	portBindings := nat.PortMap{}
	if config.Port > 0 {
		portBindings["5000/tcp"] = []nat.PortBinding{
			{
				HostIP:   "127.0.0.1",
				HostPort: strconv.Itoa(config.Port),
			},
		}
	}

	return &container.HostConfig{
		PortBindings: portBindings,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeVolume,
				Source: volumeName,
				Target: "/var/lib/registry",
			},
		},
	}
}

func (rm *RegistryManager) buildNetworkConfig(config RegistryConfig) *network.NetworkingConfig {
	if config.NetworkName == "" {
		return nil
	}

	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			config.NetworkName: {},
		},
	}
}

func (rm *RegistryManager) resolveVolumeName(config RegistryConfig) string {
	if config.VolumeName != "" {
		return config.VolumeName
	}

	return NormalizeVolumeName(config.Name)
}

// NormalizeVolumeName trims registry names and removes distribution prefixes such as kind- or k3d-.
func NormalizeVolumeName(registryName string) string {
	trimmed := strings.TrimSpace(registryName)
	if trimmed == "" {
		return ""
	}

	if strings.HasPrefix(trimmed, "kind-") || strings.HasPrefix(trimmed, "k3d-") {
		if idx := strings.Index(trimmed, "-"); idx >= 0 && idx < len(trimmed)-1 {
			candidate := trimmed[idx+1:]
			if candidate != "" {
				return candidate
			}
		}
	}

	return trimmed
}

// addClusterLabel is a no-op with network-based tracking.
// Previously used for label-based tracking, now replaced by network connections.
// Kept for interface compatibility but may be removed in future refactoring.
func (rm *RegistryManager) addClusterLabel(
	_ context.Context,
	_, _ string,
) error {
	return nil
}

func (rm *RegistryManager) stopRegistryContainer(
	ctx context.Context,
	registry container.Summary,
) error {
	if !strings.EqualFold(registry.State, "running") {
		return nil
	}

	err := rm.client.ContainerStop(ctx, registry.ID, container.StopOptions{})
	if err != nil {
		return fmt.Errorf("failed to stop registry container: %w", err)
	}

	return nil
}

func deriveRegistryVolumeName(registry container.Summary, fallback string) string {
	for _, mountPoint := range registry.Mounts {
		if mountPoint.Type == mount.TypeVolume && mountPoint.Name != "" {
			return mountPoint.Name
		}
	}

	if sanitized := NormalizeVolumeName(fallback); sanitized != "" {
		return sanitized
	}

	return strings.TrimSpace(fallback)
}

func inspectContainer(
	ctx context.Context,
	dockerClient client.APIClient,
	containerID string,
) (container.InspectResponse, error) {
	inspect, err := dockerClient.ContainerInspect(ctx, containerID)
	if err != nil {
		return container.InspectResponse{}, fmt.Errorf(
			"failed to inspect registry container: %w",
			err,
		)
	}

	return inspect, nil
}

func disconnectRegistryNetwork(
	ctx context.Context,
	dockerClient client.APIClient,
	containerID string,
	name string,
	network string,
	inspect container.InspectResponse,
) (container.InspectResponse, error) {
	if network == "" {
		return inspect, nil
	}

	err := dockerClient.NetworkDisconnect(ctx, network, containerID, true)
	//nolint:staticcheck // client.IsErrNotFound avoids importing containerd errdefs, which depguard forbids
	if err != nil && !client.IsErrNotFound(err) {
		return container.InspectResponse{}, fmt.Errorf(
			"failed to disconnect registry %s from network %s: %w",
			name,
			network,
			err,
		)
	}

	return inspectContainer(ctx, dockerClient, containerID)
}

func cleanupRegistryVolume(
	ctx context.Context,
	dockerClient client.APIClient,
	registryContainer container.Summary,
	fallbackName string,
	deleteVolume bool,
) error {
	if !deleteVolume {
		return nil
	}

	volumeName := deriveRegistryVolumeName(registryContainer, fallbackName)
	if volumeName == "" {
		return nil
	}

	err := dockerClient.VolumeRemove(ctx, volumeName, false)
	if err != nil {
		return fmt.Errorf("failed to remove registry volume: %w", err)
	}

	return nil
}

func registryAttachedToOtherClusters(
	inspect container.InspectResponse,
	ignoredNetwork string,
) bool {
	if inspect.NetworkSettings == nil || len(inspect.NetworkSettings.Networks) == 0 {
		return false
	}

	ignored := strings.ToLower(strings.TrimSpace(ignoredNetwork))

	for name := range inspect.NetworkSettings.Networks {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		lower := strings.ToLower(trimmed)
		if ignored != "" && lower == ignored {
			continue
		}

		if isClusterNetworkName(lower) {
			return true
		}
	}

	return false
}

func isClusterNetworkName(network string) bool {
	switch {
	case network == "":
		return false
	case network == "kind":
		return true
	case strings.HasPrefix(network, "kind-"):
		return true
	case network == "k3d":
		return true
	case strings.HasPrefix(network, "k3d-"):
		return true
	default:
		return false
	}
}
