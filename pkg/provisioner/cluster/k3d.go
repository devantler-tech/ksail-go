package clusterprovisioner

import (
	"context"
	"slices"

	"github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/config"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClusterProvisioner implements provisioning for k3d clusters.
type K3dClusterProvisioner struct {
	simpleCfg *v1alpha5.SimpleConfig
}

var _ ClusterProvisioner = (*K3dClusterProvisioner)(nil)

// Create provisions a k3d cluster using the loaded SimpleConfig.
func (k *K3dClusterProvisioner) Create(name string) error {
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	// Ensure name in SimpleConfig; default to ksail name
	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	k.simpleCfg.Name = target

	// Transform SimpleConfig -> ClusterConfig
	clusterCfg, err := config.TransformSimpleToClusterConfig(ctx, runtime, *k.simpleCfg, "k3d.yaml")
	if err != nil {
		return err
	}

	// Default kubeconfig options similar to CLI
	clusterCfg.KubeconfigOpts.UpdateDefaultKubeconfig = true
	clusterCfg.KubeconfigOpts.SwitchCurrentContext = true

	// Run full create sequence
	if err := client.ClusterRun(ctx, runtime, clusterCfg); err != nil {
		return err
	}

	return nil
}

// Delete tears down a k3d cluster.
func (k *K3dClusterProvisioner) Delete(name string) error {
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	cluster := &types.Cluster{Name: target}

	return client.ClusterDelete(ctx, runtime, cluster, types.ClusterDeleteOpts{})
}

// Start starts an existing k3d cluster.
func (k *K3dClusterProvisioner) Start(name string) error {
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	c, err := client.ClusterGet(ctx, runtime, &types.Cluster{Name: target})
	if err != nil {
		return err
	}

	return client.ClusterStart(ctx, runtime, c, types.ClusterStartOpts{})
}

// Stop stops a running k3d cluster.
func (k *K3dClusterProvisioner) Stop(name string) error {
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	c, err := client.ClusterGet(ctx, runtime, &types.Cluster{Name: target})
	if err != nil {
		return err
	}

	return client.ClusterStop(ctx, runtime, c)
}

// List returns cluster names managed by k3d.
func (k *K3dClusterProvisioner) List() ([]string, error) {
	ctx := context.Background()
	runtime := runtimes.SelectedRuntime

	clusters, err := client.ClusterList(ctx, runtime)
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(clusters))
	for _, c := range clusters {
		out = append(out, c.Name)
	}

	return out, nil
}

// Exists returns whether the ksail cluster name exists in k3d.
func (k *K3dClusterProvisioner) Exists(name string) (bool, error) {
	clusters, err := k.List()
	if err != nil {
		return false, err
	}

	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	return slices.Contains(clusters, target), nil
}

// NewK3dClusterProvisioner constructs a k3d provisioner instance.
func NewK3dClusterProvisioner(simpleCfg *v1alpha5.SimpleConfig) *K3dClusterProvisioner {
	return &K3dClusterProvisioner{simpleCfg: simpleCfg}
}
