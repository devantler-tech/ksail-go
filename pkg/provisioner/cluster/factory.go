package clusterprovisioner

import (
	"context"
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ErrUnsupportedDistribution is returned when an unsupported distribution is specified.
var ErrUnsupportedDistribution = errors.New("unsupported distribution")

const defaultKubeconfigPath = "~/.kube/config"

// CreateClusterProvisioner creates a cluster provisioner and returns the provisioner alongside the
// cluster name resolved from the distribution configuration.
//
//nolint:ireturn // Factory function must return interface for flexibility
func CreateClusterProvisioner(
	_ context.Context,
	distribution v1alpha1.Distribution,
	distributionConfigPath string,
	kubeconfigPath string,
) (ClusterProvisioner, string, error) {
	switch distribution {
	case v1alpha1.DistributionKind:
		return createKindProvisionerWithName(distributionConfigPath, kubeconfigPath)
	case v1alpha1.DistributionK3d:
		return createK3dProvisionerWithName(distributionConfigPath)
	default:
		return nil, "", fmt.Errorf("%w: %s", ErrUnsupportedDistribution, distribution)
	}
}

func createKindProvisionerWithName(
	distributionConfigPath string,
	kubeconfigPath string,
) (*kindprovisioner.KindClusterProvisioner, string, error) {
	kindConfigMgr := kindconfigmanager.NewConfigManager(distributionConfigPath)

	kindConfig, err := kindConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load Kind configuration: %w", err)
	}

	provisioner, err := createKindProvisionerFromConfig(kindConfig, kubeconfigPath)
	if err != nil {
		return nil, "", err
	}

	return provisioner, kindConfig.Name, nil
}

func createKindProvisionerFromConfig(
	kindConfig *v1alpha4.Cluster,
	kubeconfigPath string,
) (*kindprovisioner.KindClusterProvisioner, error) {
	provider := kindprovisioner.NewDefaultKindProviderAdapter()

	dockerClient, err := kindprovisioner.NewDefaultDockerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	if kubeconfigPath == "" {
		kubeconfigPath = defaultKubeconfigPath
	}

	return kindprovisioner.NewKindClusterProvisioner(
		kindConfig,
		kubeconfigPath,
		provider,
		dockerClient,
	), nil
}

func createK3dProvisionerWithName(
	distributionConfigPath string,
) (*k3dprovisioner.K3dClusterProvisioner, string, error) {
	k3dConfigMgr := k3dconfigmanager.NewConfigManager(distributionConfigPath)

	k3dConfig, err := k3dConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to load K3d configuration: %w", err)
	}

	provisioner := createK3dProvisionerFromConfig(k3dConfig)

	return provisioner, k3dConfig.Name, nil
}

func createK3dProvisionerFromConfig(
	k3dConfig *k3dv1alpha5.SimpleConfig,
) *k3dprovisioner.K3dClusterProvisioner {
	clientProvider := k3dprovisioner.NewDefaultK3dClientAdapter()
	configProvider := k3dprovisioner.NewDefaultK3dConfigAdapter()

	return k3dprovisioner.NewK3dClusterProvisioner(
		k3dConfig,
		clientProvider,
		configProvider,
	)
}
