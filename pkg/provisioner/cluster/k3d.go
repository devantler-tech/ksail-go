package clusterprovisioner

import (
	"context"
	"fmt"
	"slices"

	"github.com/k3d-io/k3d/v5/pkg/client"
	"github.com/k3d-io/k3d/v5/pkg/config"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
)

// K3dClusterProvisioner implements provisioning for k3d clusters.
type K3dClusterProvisioner struct {
	config *v1alpha5.SimpleConfig
}

var _ ClusterProvisioner = (*K3dClusterProvisioner)(nil)

// Allow stubbing in tests by routing calls through variables.
var (
	k3dClusterRun  = client.ClusterRun
	k3dClusterGet  = client.ClusterGet
	k3dClusterStop = client.ClusterStop
	k3dClusterStart = client.ClusterStart
	k3dClusterDelete = client.ClusterDelete
	k3dClusterList = client.ClusterList
	k3dTransformSimpleToClusterConfig = config.TransformSimpleToClusterConfig
)

// NewK3dClusterProvisioner constructs a k3d provisioner instance using only the k3d SimpleConfig.
func NewK3dClusterProvisioner(simpleCfg *v1alpha5.SimpleConfig) *K3dClusterProvisioner {
	return &K3dClusterProvisioner{config: simpleCfg}
}

// Create provisions a k3d cluster using the loaded SimpleConfig.
func (k *K3dClusterProvisioner) Create(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	// Ensure name in SimpleConfig; default to ksail name
	target := k.resolveName(name)

	k.config.Name = target

	// Transform SimpleConfig -> ClusterConfig
	clusterCfg, err := k3dTransformSimpleToClusterConfig(ctx, runtime, *k.config, "k3d.yaml")
	if err != nil {
		return fmt.Errorf("failed to transform simple config to cluster config: %w", err)
	}

	// Default kubeconfig options similar to CLI
	clusterCfg.KubeconfigOpts.UpdateDefaultKubeconfig = true
	clusterCfg.KubeconfigOpts.SwitchCurrentContext = true

	// Run full create sequence
	err = k3dClusterRun(ctx, runtime, clusterCfg)
	if err != nil {
		return fmt.Errorf("failed to run k3d cluster: %w", err)
	}

	return nil
}

// Delete tears down a k3d cluster.
func (k *K3dClusterProvisioner) Delete(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	target := k.resolveName(name)

	cluster := &types.Cluster{Name: target}

	err := k3dClusterDelete(ctx, runtime, cluster, types.ClusterDeleteOpts{
		SkipRegistryCheck: false,
	})
	if err != nil {
		return fmt.Errorf("failed to delete k3d cluster %q: %w", target, err)
	}

	return nil
}

// Start starts an existing k3d cluster.
func (k *K3dClusterProvisioner) Start(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	target := k.resolveName(name)

	c, err := k3dClusterGet(ctx, runtime, &types.Cluster{Name: target})
	if err != nil {
		return fmt.Errorf("failed to get k3d cluster %q: %w", target, err)
	}

	err = k3dClusterStart(ctx, runtime, c, types.ClusterStartOpts{
		WaitForServer:   false,
		Timeout:         0,
		NodeHooks:       nil,
		EnvironmentInfo: nil,
		Intent:          "",
		HostAliases:     nil,
	})
	if err != nil {
		return fmt.Errorf("failed to start k3d cluster %q: %w", target, err)
	}

	return nil
}

// Stop stops a running k3d cluster.
func (k *K3dClusterProvisioner) Stop(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	target := k.resolveName(name)

	c, err := k3dClusterGet(ctx, runtime, &types.Cluster{Name: target})
	if err != nil {
		return fmt.Errorf("failed to get k3d cluster %q: %w", target, err)
	}

	err = k3dClusterStop(ctx, runtime, c)
	if err != nil {
		return fmt.Errorf("failed to stop k3d cluster %q: %w", target, err)
	}

	return nil
}

// List returns cluster names managed by k3d.
func (k *K3dClusterProvisioner) List(ctx context.Context) ([]string, error) {
	rt := runtimes.SelectedRuntime

	clusters, err := k3dClusterList(ctx, rt)
	if err != nil {
		return nil, fmt.Errorf("failed to list k3d clusters: %w", err)
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
		return false, err
	}

	target := k.resolveName(name)

	return slices.Contains(clusters, target), nil
}

// --- internals ---

// resolveName determines the effective cluster name to operate on.
// If the provided name is empty, it falls back to the provisioner's configured name.
func (k *K3dClusterProvisioner) resolveName(name string) string {
	if name != "" {
		return name
	}

	return k.config.Name
}
