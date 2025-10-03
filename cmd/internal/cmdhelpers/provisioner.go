package cmdhelpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ErrUnsupportedDistribution is returned when an unsupported distribution is specified.
var ErrUnsupportedDistribution = errors.New("unsupported distribution")

// CreateClusterProvisionerWithName creates the appropriate provisioner based on the cluster distribution
// and returns both the provisioner and the cluster name from the distribution config.
// This function loads the distribution config once and extracts both the provisioner and name efficiently.
//
//nolint:ireturn // Factory function must return interface for flexibility
func CreateClusterProvisionerWithName(
	_ context.Context,
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, string, error) {
	switch cluster.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return createKindProvisionerWithName(cluster)
	case v1alpha1.DistributionK3d:
		return createK3dProvisionerWithName(cluster)
	default:
		return nil, "", fmt.Errorf("%w: %s", ErrUnsupportedDistribution, cluster.Spec.Distribution)
	}
}

// createKindProvisionerWithName creates a Kind cluster provisioner and returns the cluster name.
//
//nolint:ireturn // Factory function must return interface for flexibility
func createKindProvisionerWithName(
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, string, error) {
	// Load Kind configuration using config manager (only once)
	kindConfigMgr := kindconfigmanager.NewConfigManager(cluster.Spec.DistributionConfig)

	kindConfig, err := kindConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load Kind configuration: %w", err)
	}

	provisioner, err := createKindProvisionerFromConfig(cluster, kindConfig)
	if err != nil {
		return nil, "", err
	}

	return provisioner, kindConfig.Name, nil
}

// createKindProvisionerFromConfig creates a Kind cluster provisioner from an already-loaded config.
//
//nolint:ireturn // Factory function must return interface for flexibility
func createKindProvisionerFromConfig(
	cluster *v1alpha1.Cluster,
	kindConfig *v1alpha4.Cluster,
) (clusterprovisioner.ClusterProvisioner, error) {
	// Create Kind provider adapter
	provider := kindprovisioner.NewDefaultKindProviderAdapter()

	// Create Docker client
	dockerClient, err := kindprovisioner.NewDefaultDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Determine kubeconfig path
	kubeconfig := cluster.Spec.Connection.Kubeconfig
	if kubeconfig == "" {
		kubeconfig = "~/.kube/config"
	}

	return kindprovisioner.NewKindClusterProvisioner(
		kindConfig,
		kubeconfig,
		provider,
		dockerClient,
	), nil
}

// createK3dProvisionerWithName creates a K3d cluster provisioner and returns the cluster name.
//
//nolint:ireturn // Factory function must return interface for flexibility
func createK3dProvisionerWithName(
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, string, error) {
	// Load K3d configuration using config manager (only once)
	k3dConfigMgr := k3dconfigmanager.NewConfigManager(cluster.Spec.DistributionConfig)

	k3dConfig, err := k3dConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load K3d configuration: %w", err)
	}

	provisioner := createK3dProvisionerFromConfig(k3dConfig)

	return provisioner, k3dConfig.Name, nil
}

// createK3dProvisionerFromConfig creates a K3d cluster provisioner from an already-loaded config.
//
//nolint:ireturn // Factory function must return interface for flexibility
func createK3dProvisionerFromConfig(
	k3dConfig *k3dv1alpha5.SimpleConfig,
) clusterprovisioner.ClusterProvisioner {
	// Create K3d client and config adapters
	clientProvider := k3dprovisioner.NewDefaultK3dClientAdapter()
	configProvider := k3dprovisioner.NewDefaultK3dConfigAdapter()

	return k3dprovisioner.NewK3dClusterProvisioner(
		k3dConfig,
		clientProvider,
		configProvider,
	)
}
