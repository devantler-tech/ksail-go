package clusterprovisioner

import (
	"context"
	"fmt"
	"slices"
	"time"

	pathutils "github.com/devantler-tech/ksail-go/internal/utils/path"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// KindClusterProvisioner is an implementation of the ClusterProvisioner interface for provisioning kind clusters.
type KindClusterProvisioner struct {
	kubeConfig     string
	kindConfig     *v1alpha4.Cluster
	dockerProvider *cluster.Provider
}

// NewKindClusterProvisioner creates a new KindClusterProvisioner.
func NewKindClusterProvisioner(kindConfig *v1alpha4.Cluster, kubeConfig string) *KindClusterProvisioner {
	return &KindClusterProvisioner{
		kubeConfig: kubeConfig,
		kindConfig: kindConfig,
		dockerProvider: cluster.NewProvider(
			cluster.ProviderWithLogger(kindcmd.NewLogger()),
		),
	}
}

// Create creates a kind cluster.
func (k *KindClusterProvisioner) Create(name string) error {
	target := setName(name, k.kindConfig.Name)

	err := k.dockerProvider.Create(
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
func (k *KindClusterProvisioner) Delete(name string) error {
	target := setName(name, k.kindConfig.Name)

	kubeconfigPath, err := pathutils.ExpandPath(k.kubeConfig)
	if err != nil {
		return fmt.Errorf("failed to expand kubeconfig path: %w", err)
	}

	err = k.dockerProvider.Delete(target, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to delete kind cluster: %w", err)
	}

	return nil
}

// Start starts a kind cluster.
func (k *KindClusterProvisioner) Start(name string) error {
	const dockerStartTimeout = 30 * time.Second

	target := setName(name, k.kindConfig.Name)

	nodes, err := k.dockerProvider.ListNodes(target)
	if err != nil {
		return fmt.Errorf("cluster '%s': %w", target, err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("cluster '%s': no nodes found", target)
	}

	// Create a Docker client
	ctx, cancel := context.WithTimeout(context.Background(), dockerStartTimeout)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	defer func() { _ = cli.Close() }()

	for _, n := range nodes {
		// Start each node container by name using Docker SDK
		name := n.String()

		err := cli.ContainerStart(ctx, name, container.StartOptions{
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
func (k *KindClusterProvisioner) Stop(name string) error {
	const dockerStopTimeout = 60 * time.Second

	target := setName(name, k.kindConfig.Name)

	nodes, err := k.dockerProvider.ListNodes(target)
	if err != nil {
		return fmt.Errorf("failed to list nodes for cluster '%s': %w", target, err)
	}

	if len(nodes) == 0 {
		return fmt.Errorf("cluster '%s' not found", target)
	}

	// Create a Docker client
	ctx, cancel := context.WithTimeout(context.Background(), dockerStopTimeout)
	defer cancel()

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to create docker client: %w", err)
	}

	defer func() { _ = cli.Close() }()

	for _, n := range nodes {
		// Stop each node container by name using Docker SDK
		name := n.String()
		// Graceful stop with default timeout
		err := cli.ContainerStop(ctx, name, container.StopOptions{
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
func (k *KindClusterProvisioner) List() ([]string, error) {
	clusters, err := k.dockerProvider.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	return clusters, nil
}

// Exists checks if a kind cluster exists.
func (k *KindClusterProvisioner) Exists(name string) (bool, error) {
	clusters, err := k.dockerProvider.List()
	if err != nil {
		return false, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	target := setName(name, k.kindConfig.Name)

	if slices.Contains(clusters, target) {
		return true, nil
	}

	return false, nil
}

// --- internals ---

func setName(name string, kindConfigName string) string {
	target := name
	if target == "" {
		target = kindConfigName
	}

	return target
}
