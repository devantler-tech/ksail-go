// Package kindprovisioner provides implementations of the Provisioner interface
// for provisioning clusters in different providers.
package kindprovisioner

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	iopath "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/registry"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
)

// ErrClusterNotFound is returned when a cluster is not found.
var ErrClusterNotFound = errors.New("cluster not found")

// KindProvider describes the subset of methods from kind's Provider used here.
type KindProvider interface {
	Create(name string, opts ...cluster.CreateOption) error
	Delete(name, kubeconfigPath string) error
	List() ([]string, error)
	ListNodes(name string) ([]string, error)
}

// KindClusterProvisioner is an implementation of the ClusterProvisioner interface for provisioning kind clusters.
type KindClusterProvisioner struct {
	kubeConfig      string
	kindConfig      *v1alpha4.Cluster
	provider        KindProvider
	client          client.APIClient
	registryManager *registry.Manager
}

// NewKindClusterProvisioner constructs a KindClusterProvisioner with explicit dependencies
// for the kind provider and docker client. This supports both production wiring
// and unit testing via mocks.
func NewKindClusterProvisioner(
	kindConfig *v1alpha4.Cluster,
	kubeConfig string,
	provider KindProvider,
	dockerClient client.APIClient,
) *KindClusterProvisioner {
	return &KindClusterProvisioner{
		kubeConfig:      kubeConfig,
		kindConfig:      kindConfig,
		provider:        provider,
		client:          dockerClient,
		registryManager: registry.NewManager(dockerClient),
	}
}

// Create creates a kind cluster and any configured mirror registries.
func (k *KindClusterProvisioner) Create(ctx context.Context, name string) error {
	// First, create any required mirror registries
	err := k.createMirrorRegistries(ctx)
	if err != nil {
		return fmt.Errorf("failed to create mirror registries: %w", err)
	}

	// Then create the Kind cluster
	target := setName(name, k.kindConfig.Name)

	err = k.provider.Create(
		target,
		cluster.CreateWithV1Alpha4Config(k.kindConfig),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
	if err != nil {
		return fmt.Errorf("failed to create kind cluster: %w", err)
	}

	return nil
}

// Delete deletes a kind cluster.
func (k *KindClusterProvisioner) Delete(_ context.Context, name string) error {
	target := setName(name, k.kindConfig.Name)

	kubeconfigPath, _ := iopath.ExpandHomePath(k.kubeConfig)

	err := k.provider.Delete(target, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w", err)
	}

	return nil
}

// Start starts a kind cluster.
func (k *KindClusterProvisioner) Start(ctx context.Context, name string) error {
	const dockerStartTimeout = 30 * time.Second

	target := setName(name, k.kindConfig.Name)

	nodes, err := k.provider.ListNodes(target)
	if err != nil {
		return fmt.Errorf("cluster '%s': %w", target, err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("%w", ErrClusterNotFound)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, dockerStartTimeout)
	defer cancel()

	for _, name := range nodes {
		// Start each node container by name using Docker SDK
		err := k.client.ContainerStart(timeoutCtx, name, container.StartOptions{
			CheckpointID:  "",
			CheckpointDir: "",
		})
		if err != nil {
			return fmt.Errorf("docker start failed for %s: %w", name, err)
		}
	}

	return nil
}

// Stop stops a kind cluster.
func (k *KindClusterProvisioner) Stop(ctx context.Context, name string) error {
	const dockerStopTimeout = 60 * time.Second

	target := setName(name, k.kindConfig.Name)

	nodes, err := k.provider.ListNodes(target)
	if err != nil {
		return fmt.Errorf("failed to list nodes for cluster '%s': %w", target, err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("%w", ErrClusterNotFound)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, dockerStopTimeout)
	defer cancel()

	for _, name := range nodes {
		// Stop each node container by name using Docker SDK
		// Graceful stop with default timeout
		err := k.client.ContainerStop(timeoutCtx, name, container.StopOptions{
			Signal:  "",
			Timeout: nil,
		})
		if err != nil {
			return fmt.Errorf("docker stop failed for %s: %w", name, err)
		}
	}

	return nil
}

// List returns all kind clusters.
func (k *KindClusterProvisioner) List(_ context.Context) ([]string, error) {
	clusters, err := k.provider.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	return clusters, nil
}

// Exists checks if a kind cluster exists.
func (k *KindClusterProvisioner) Exists(_ context.Context, name string) (bool, error) {
	clusters, err := k.provider.List()
	if err != nil {
		return false, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	target := setName(name, k.kindConfig.Name)

	if slices.Contains(clusters, target) {
		return true, nil
	}

	return false, nil
}

// createMirrorRegistries extracts and creates mirror registries from Kind config.
func (k *KindClusterProvisioner) createMirrorRegistries(ctx context.Context) error {
	registries, err := registry.ExtractRegistriesFromKind(k.kindConfig)
	if err != nil {
		return fmt.Errorf("failed to extract registries: %w", err)
	}

	for _, reg := range registries {
		err := k.registryManager.CreateRegistry(ctx, reg)
		if err != nil {
			return fmt.Errorf("failed to create registry %s: %w", reg.Name, err)
		}
	}

	return nil
}

// --- internals ---

func setName(name string, kindConfigName string) string {
	target := name
	if target == "" {
		target = kindConfigName
	}

	return target
}
