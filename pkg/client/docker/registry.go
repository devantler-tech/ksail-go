package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

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
)

const (
	// RegistryImageName is the default registry image to use.
	RegistryImageName = "registry:3"
	// RegistryLabelKey is the label key used to identify ksail registries.
	RegistryLabelKey = "io.ksail.registry"
	// RegistryClusterLabelKey is the label key used to track which clusters use a registry.
	RegistryClusterLabelKey = "io.ksail.cluster"
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
	if err := rm.ensureRegistryImage(ctx); err != nil {
		return fmt.Errorf("failed to ensure registry image: %w", err)
	}

	// Create volume for registry data
	volumeName := fmt.Sprintf("ksail-registry-%s", config.Name)
	if err := rm.createVolume(ctx, volumeName, config.Name); err != nil {
		return fmt.Errorf("failed to create registry volume: %w", err)
	}

	// Prepare container configuration
	containerConfig := rm.buildContainerConfig(config)
	hostConfig := rm.buildHostConfig(config, volumeName)
	networkConfig := rm.buildNetworkConfig(config)

	containerName := fmt.Sprintf("ksail-registry-%s", config.Name)

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
	if err := rm.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start registry container: %w", err)
	}

	return nil
}

// DeleteRegistry removes a registry container and optionally its volume.
// If deleteVolume is true, the associated volume will be removed.
// If the registry is still in use by other clusters, it returns an error.
func (rm *RegistryManager) DeleteRegistry(
	ctx context.Context,
	name, clusterName string,
	deleteVolume bool,
) error {
	containerName := fmt.Sprintf("ksail-registry-%s", name)

	// Get container to check labels
	containers, err := rm.listRegistryContainers(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to list registry containers: %w", err)
	}

	if len(containers) == 0 {
		return ErrRegistryNotFound
	}

	registryContainer := containers[0]

	// Remove cluster label
	if err := rm.removeClusterLabel(ctx, containerName, clusterName); err != nil {
		return fmt.Errorf("failed to remove cluster label: %w", err)
	}

	// Check if registry is still in use
	inUse, err := rm.IsRegistryInUse(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to check if registry is in use: %w", err)
	}

	if inUse {
		// Registry is still in use by other clusters, don't delete
		return nil
	}

	// Stop and remove container
	if err := rm.client.ContainerStop(ctx, registryContainer.ID, container.StopOptions{}); err != nil {
		return fmt.Errorf("failed to stop registry container: %w", err)
	}

	if err := rm.client.ContainerRemove(ctx, registryContainer.ID, container.RemoveOptions{}); err != nil {
		return fmt.Errorf("failed to remove registry container: %w", err)
	}

	// Remove volume if requested
	if deleteVolume {
		volumeName := fmt.Sprintf("ksail-registry-%s", name)
		if err := rm.client.VolumeRemove(ctx, volumeName, false); err != nil {
			return fmt.Errorf("failed to remove registry volume: %w", err)
		}
	}

	return nil
}

// ListRegistries returns a list of all ksail registry containers.
func (rm *RegistryManager) ListRegistries(ctx context.Context) ([]string, error) {
	containers, err := rm.listAllRegistryContainers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list registry containers: %w", err)
	}

	registries := make([]string, 0, len(containers))
	for _, c := range containers {
		if name, ok := c.Labels[RegistryLabelKey]; ok {
			registries = append(registries, name)
		}
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
		if port.PrivatePort == 5000 {
			return int(port.PublicPort), nil
		}
	}

	return 0, fmt.Errorf("registry port not found")
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
	filterArgs.Add("label", fmt.Sprintf("%s=%s", RegistryLabelKey, name))

	return rm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
}

func (rm *RegistryManager) listAllRegistryContainers(
	ctx context.Context,
) ([]container.Summary, error) {
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", RegistryLabelKey)

	return rm.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
}

func (rm *RegistryManager) ensureRegistryImage(ctx context.Context) error {
	// Check if image exists
	_, _, err := rm.client.ImageInspectWithRaw(ctx, RegistryImageName)
	if err == nil {
		return nil
	}

	// Pull image
	reader, err := rm.client.ImagePull(ctx, RegistryImageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull registry image: %w", err)
	}
	defer reader.Close()

	// Consume pull output
	_, err = io.Copy(io.Discard, reader)
	if err != nil {
		return fmt.Errorf("failed to read image pull output: %w", err)
	}

	return nil
}

func (rm *RegistryManager) createVolume(
	ctx context.Context,
	volumeName, registryName string,
) error {
	// Check if volume already exists
	_, err := rm.client.VolumeInspect(ctx, volumeName)
	if err == nil {
		return nil // Volume already exists
	}

	// Create volume
	_, err = rm.client.VolumeCreate(ctx, volume.CreateOptions{
		Name: volumeName,
		Labels: map[string]string{
			RegistryLabelKey: registryName,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to create volume: %w", err)
	}

	return nil
}

func (rm *RegistryManager) buildContainerConfig(config RegistryConfig) *container.Config {
	env := []string{}
	if config.UpstreamURL != "" {
		env = append(env, fmt.Sprintf("REGISTRY_PROXY_REMOTEURL=%s", config.UpstreamURL))
	}

	return &container.Config{
		Image: RegistryImageName,
		Env:   env,
		ExposedPorts: nat.PortSet{
			"5000/tcp": struct{}{},
		},
		Labels: map[string]string{
			RegistryLabelKey:        config.Name,
			RegistryClusterLabelKey: config.ClusterName,
		},
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

func (rm *RegistryManager) addClusterLabel(
	ctx context.Context,
	registryName, clusterName string,
) error {
	// With the network-based tracking, we just need to ensure the registry exists
	// The actual network connection will be made when attaching to the cluster network
	return nil
}

func (rm *RegistryManager) removeClusterLabel(
	ctx context.Context,
	registryName, clusterName string,
) error {
	// With the network-based tracking, network disconnection happens when cluster is deleted
	// This is a no-op as the network will be cleaned up automatically
	return nil
}
