package clusterprovisioner

import (
	"context"
	"fmt"
	"slices"

	"github.com/devantler-tech/ksail-go/internal/utils"
	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// KindClusterProvisioner is an implementation of the ClusterProvisioner interface for provisioning kind clusters.
type KindClusterProvisioner struct {
	ksailConfig    *ksailcluster.Cluster
	kindConfig     *v1alpha4.Cluster
	dockerProvider *cluster.Provider
	dockerClient   *client.Client
}

// Create creates a kind cluster.
func (k *KindClusterProvisioner) Create(name string) error {
	target := name
	if target == "" {
		target = k.ksailConfig.Metadata.Name
	}
	return k.dockerProvider.Create(
		target,
		cluster.CreateWithV1Alpha4Config(k.kindConfig),
		cluster.CreateWithDisplayUsage(true),
		cluster.CreateWithDisplaySalutation(true),
	)
}

// Delete deletes a kind cluster.
func (k *KindClusterProvisioner) Delete(name string) error {
	target := name
	if target == "" {
		target = k.ksailConfig.Metadata.Name
	}
	kubeconfigPath, err := utils.ExpandPath(k.ksailConfig.Spec.Connection.Kubeconfig)
	if err != nil {
		return err
	}
	return k.dockerProvider.Delete(target, kubeconfigPath)
}

// Starts a kind cluster.
func (k *KindClusterProvisioner) Start(name string) error {
	target := name
	if target == "" {
		target = k.ksailConfig.Metadata.Name
	}
	nodes, err := k.dockerProvider.ListNodes(target)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("cluster '%s' not found", target)
	}

	ctx := context.Background()
	for _, n := range nodes {
		// Start each node container using Docker SDK
		if err := k.dockerClient.ContainerStart(ctx, n.String(), container.StartOptions{}); err != nil {
			return fmt.Errorf("failed to start container %s: %v", n.String(), err)
		}
	}
	return nil
}

// Stops a kind cluster.
func (k *KindClusterProvisioner) Stop(name string) error {
	target := name
	if target == "" {
		target = k.ksailConfig.Metadata.Name
	}
	nodes, err := k.dockerProvider.ListNodes(target)
	if err != nil {
		return err
	}
	if len(nodes) == 0 {
		return fmt.Errorf("cluster '%s' not found", target)
	}

	ctx := context.Background()
	for _, n := range nodes {
		// Stop each node container using Docker SDK
		timeout := 30 // 30 seconds timeout
		if err := k.dockerClient.ContainerStop(ctx, n.String(), container.StopOptions{Timeout: &timeout}); err != nil {
			return fmt.Errorf("failed to stop container %s: %v", n.String(), err)
		}
	}
	return nil
}

// Lists all kind clusters.
func (k *KindClusterProvisioner) List() ([]string, error) {
	return k.dockerProvider.List()
}

// Checks if a kind cluster exists.
func (k *KindClusterProvisioner) Exists(name string) (bool, error) {
	clusters, err := k.dockerProvider.List()
	if err != nil {
		return false, err
	}
	target := name
	if target == "" {
		target = k.ksailConfig.Metadata.Name
	}
	if slices.Contains(clusters, target) {
		return true, nil
	}
	return false, nil
}

// NewKindClusterProvisioner creates a new KindClusterProvisioner.
func NewKindClusterProvisioner(ksailConfig *ksailcluster.Cluster, kindConfig *v1alpha4.Cluster) *KindClusterProvisioner {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		// Fall back to nil client, which will cause operations to fail gracefully
		dockerClient = nil
	}

	return &KindClusterProvisioner{
		ksailConfig: ksailConfig,
		kindConfig:  kindConfig,
		dockerProvider: cluster.NewProvider(
			cluster.ProviderWithLogger(kindcmd.NewLogger()),
		),
		dockerClient: dockerClient,
	}
}
