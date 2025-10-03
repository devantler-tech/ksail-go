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
)

// ErrUnsupportedDistribution is returned when an unsupported distribution is specified.
var ErrUnsupportedDistribution = errors.New("unsupported distribution")

// GetClusterNameFromConfig extracts the cluster name from the distribution configuration.
func GetClusterNameFromConfig(cluster *v1alpha1.Cluster) (string, error) {
	switch cluster.Spec.Distribution {
	case v1alpha1.DistributionKind:
		kindConfigMgr := kindconfigmanager.NewConfigManager(cluster.Spec.DistributionConfig)

		kindConfig, err := kindConfigMgr.LoadConfig(nil)
		if err != nil {
			return "", fmt.Errorf("failed to load Kind configuration: %w", err)
		}

		return kindConfig.Name, nil
	case v1alpha1.DistributionK3d:
		k3dConfigMgr := k3dconfigmanager.NewConfigManager(cluster.Spec.DistributionConfig)

		k3dConfig, err := k3dConfigMgr.LoadConfig(nil)
		if err != nil {
			return "", fmt.Errorf("failed to load K3d configuration: %w", err)
		}

		return k3dConfig.Name, nil
	default:
		return "", fmt.Errorf("%w: %s", ErrUnsupportedDistribution, cluster.Spec.Distribution)
	}
}

// CreateClusterProvisioner creates the appropriate provisioner based on the cluster distribution.
//
//nolint:ireturn // Factory function must return interface for flexibility
func CreateClusterProvisioner(
	_ context.Context,
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, error) {
	switch cluster.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return createKindProvisioner(cluster)
	case v1alpha1.DistributionK3d:
		return createK3dProvisioner(cluster)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDistribution, cluster.Spec.Distribution)
	}
}

// createKindProvisioner creates a Kind cluster provisioner.
//
//nolint:ireturn // Factory function must return interface for flexibility
func createKindProvisioner(
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, error) {
	// Load Kind configuration using config manager
	kindConfigMgr := kindconfigmanager.NewConfigManager(cluster.Spec.DistributionConfig)

	kindConfig, err := kindConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load Kind configuration: %w", err)
	}

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

// createK3dProvisioner creates a K3d cluster provisioner.
//
//nolint:ireturn // Factory function must return interface for flexibility
func createK3dProvisioner(
	cluster *v1alpha1.Cluster,
) (clusterprovisioner.ClusterProvisioner, error) {
	// Load K3d configuration using config manager
	k3dConfigMgr := k3dconfigmanager.NewConfigManager(cluster.Spec.DistributionConfig)

	k3dConfig, err := k3dConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load K3d configuration: %w", err)
	}

	// Create K3d client and config adapters
	clientProvider := k3dprovisioner.NewDefaultK3dClientAdapter()
	configProvider := k3dprovisioner.NewDefaultK3dConfigAdapter()

	return k3dprovisioner.NewK3dClusterProvisioner(
		k3dConfig,
		clientProvider,
		configProvider,
	), nil
}
