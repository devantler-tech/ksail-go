package k3dprovisioner

import (
	"context"
	"fmt"
	"slices"

	"github.com/devantler-tech/ksail-go/pkg/client/docker"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClusterProvisioner implements provisioning for k3d clusters.
type K3dClusterProvisioner struct {
	simpleCfg       *v1alpha5.SimpleConfig
	clientProvider  K3dClientProvider
	configProvider  K3dConfigProvider
	registryManager *docker.RegistryManager
}

// NewK3dClusterProvisioner constructs a k3d provisioner instance.
func NewK3dClusterProvisioner(
	simpleCfg *v1alpha5.SimpleConfig,
	clientProvider K3dClientProvider,
	configProvider K3dConfigProvider,
) *K3dClusterProvisioner {
	// Note: registry manager will be nil initially since we don't have a Docker client here
	// It would need to be injected if we want to support registry creation for K3d
	return &K3dClusterProvisioner{
		simpleCfg:       simpleCfg,
		clientProvider:  clientProvider,
		configProvider:  configProvider,
		registryManager: nil,
	}
}

// Create provisions a k3d cluster using the loaded SimpleConfig and creates any configured mirror registries.
func (k *K3dClusterProvisioner) Create(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	// Ensure name in SimpleConfig; default to ksail name
	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	k.simpleCfg.Name = target

	// K3d has native registry support, so we'll let K3d handle registry creation
	// If registries are specified in the config, K3d will create them automatically

	// Transform SimpleConfig -> ClusterConfig
	clusterCfg, err := k.configProvider.TransformSimpleToClusterConfig(
		ctx,
		runtime,
		*k.simpleCfg,
		"k3d.yaml",
	)
	if err != nil {
		return fmt.Errorf("transform simple to cluster config: %w", err)
	}

	// Default kubeconfig options similar to CLI
	clusterCfg.KubeconfigOpts.UpdateDefaultKubeconfig = true
	clusterCfg.KubeconfigOpts.SwitchCurrentContext = true

	// Run full create sequence
	err = k.clientProvider.ClusterRun(ctx, runtime, clusterCfg)
	if err != nil {
		return fmt.Errorf("cluster run: %w", err)
	}

	return nil
}

// Delete tears down a k3d cluster.
func (k *K3dClusterProvisioner) Delete(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	var cluster types.Cluster

	cluster.Name = target

	var opts types.ClusterDeleteOpts

	err := k.clientProvider.ClusterDelete(ctx, runtime, &cluster, opts)
	if err != nil {
		return fmt.Errorf("cluster delete: %w", err)
	}

	return nil
}

// Start starts an existing k3d cluster.
func (k *K3dClusterProvisioner) Start(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	var cluster types.Cluster

	cluster.Name = target

	k3dCluster, err := k.clientProvider.ClusterGet(ctx, runtime, &cluster)
	if err != nil {
		return fmt.Errorf("cluster get: %w", err)
	}

	var startOpts types.ClusterStartOpts

	err = k.clientProvider.ClusterStart(ctx, runtime, k3dCluster, startOpts)
	if err != nil {
		return fmt.Errorf("cluster start: %w", err)
	}

	return nil
}

// Stop stops a running k3d cluster.
func (k *K3dClusterProvisioner) Stop(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	var cluster types.Cluster

	cluster.Name = target

	c, err := k.clientProvider.ClusterGet(ctx, runtime, &cluster)
	if err != nil {
		return fmt.Errorf("cluster get: %w", err)
	}

	err = k.clientProvider.ClusterStop(ctx, runtime, c)
	if err != nil {
		return fmt.Errorf("cluster stop: %w", err)
	}

	return nil
}

// List returns cluster names managed by k3d.
func (k *K3dClusterProvisioner) List(ctx context.Context) ([]string, error) {
	runtime := runtimes.SelectedRuntime

	clusters, err := k.clientProvider.ClusterList(ctx, runtime)
	if err != nil {
		return nil, fmt.Errorf("cluster list: %w", err)
	}

	out := make([]string, 0, len(clusters))
	for _, c := range clusters {
		out = append(out, c.Name)
	}

	return out, nil
}

// Exists returns whether the ksail cluster name exists in k3d.
func (k *K3dClusterProvisioner) Exists(ctx context.Context, name string) (bool, error) {
	clusters, err := k.List(ctx)
	if err != nil {
		return false, fmt.Errorf("list: %w", err)
	}

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	return slices.Contains(clusters, target), nil
}
