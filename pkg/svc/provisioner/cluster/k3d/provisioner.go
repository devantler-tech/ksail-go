package k3dprovisioner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/k3d-io/k3d/v5/pkg/runtimes"
	"github.com/k3d-io/k3d/v5/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
)

// K3dClusterProvisioner implements provisioning for k3d clusters.
type K3dClusterProvisioner struct {
	simpleCfg      *v1alpha5.SimpleConfig
	clientProvider K3dClientProvider
	configProvider K3dConfigProvider
}

// NewK3dClusterProvisioner constructs a k3d provisioner instance.
func NewK3dClusterProvisioner(
	simpleCfg *v1alpha5.SimpleConfig,
	clientProvider K3dClientProvider,
	configProvider K3dConfigProvider,
) *K3dClusterProvisioner {
	return &K3dClusterProvisioner{
		simpleCfg:      simpleCfg,
		clientProvider: clientProvider,
		configProvider: configProvider,
	}
}

// Create provisions a k3d cluster using the loaded SimpleConfig.
func (k *K3dClusterProvisioner) Create(ctx context.Context, name string) error {
	runtime := runtimes.SelectedRuntime

	// Ensure name in SimpleConfig; default to ksail name
	target := name
	if target == "" {
		target = k.simpleCfg.Name
	}

	k.simpleCfg.Name = target

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

	// Ensure ~/.kube directory exists for kubeconfig writing
	err = ensureKubeDirectoryExists()
	if err != nil {
		return fmt.Errorf("ensure kube directory: %w", err)
	}

	// Run full create sequence
	err = k.clientProvider.ClusterRun(ctx, runtime, clusterCfg)
	if err != nil {
		return fmt.Errorf("cluster run: %w", err)
	}

	// Explicitly write kubeconfig after cluster creation
	// K3d's UpdateDefaultKubeconfig doesn't always work reliably, so we ensure it's written
	err = k.writeKubeconfig(ctx, runtime, target, &clusterCfg.KubeconfigOpts)
	if err != nil {
		return fmt.Errorf("write kubeconfig: %w", err)
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

// ensureKubeDirectoryExists creates the ~/.kube directory if it doesn't exist.
func ensureKubeDirectoryExists() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	kubeDir := filepath.Join(homeDir, ".kube")
	err = os.MkdirAll(kubeDir, 0o700)
	if err != nil {
		return fmt.Errorf("create kube directory: %w", err)
	}

	return nil
}

// writeKubeconfig writes the kubeconfig for a k3d cluster to the configured location.
func (k *K3dClusterProvisioner) writeKubeconfig(
	ctx context.Context,
	runtime runtimes.Runtime,
	clusterName string,
	opts *v1alpha5.SimpleConfigOptionsKubeconfig,
) error {
	// Determine if we should write the kubeconfig
	if !opts.UpdateDefaultKubeconfig {
		return nil
	}

	// Get the cluster details
	cluster := &types.Cluster{Name: clusterName}
	clusterDetails, err := k.clientProvider.ClusterGet(ctx, runtime, cluster)
	if err != nil {
		return fmt.Errorf("get cluster details: %w", err)
	}

	// Get the kubeconfig from the cluster
	kubeconfig, err := k.clientProvider.KubeconfigGet(ctx, runtime, clusterDetails)
	if err != nil {
		return fmt.Errorf("get kubeconfig: %w", err)
	}

	// Determine the output path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("get home directory: %w", err)
	}

	kubeconfigPath := filepath.Join(homeDir, ".kube", "config")

	// Write the kubeconfig to the file
	err = clientcmd.WriteToFile(*kubeconfig, kubeconfigPath)
	if err != nil {
		return fmt.Errorf("write kubeconfig to file: %w", err)
	}

	return nil
}
