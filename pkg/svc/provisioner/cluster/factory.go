package clusterprovisioner

import (
	"context"
	"errors"
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	k3dconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/k3d"
	kindconfigmanager "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/svc/provisioner/cluster/kind"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ErrUnsupportedDistribution is returned when an unsupported distribution is specified.
var ErrUnsupportedDistribution = errors.New("unsupported distribution")

const defaultKubeconfigPath = "~/.kube/config"

// Factory creates distribution-specific cluster provisioners based on the KSail cluster configuration.
type Factory interface {
	Create(ctx context.Context, cluster *v1alpha1.Cluster) (ClusterProvisioner, any, error)
}

// DefaultFactory implements Factory using the existing CreateClusterProvisioner helper.
type DefaultFactory struct{}

// Create selects the correct distribution provisioner for the KSail cluster configuration.
//
//nolint:ireturn // Factory interface intentionally returns ClusterProvisioner implementations.
func (DefaultFactory) Create(
	_ context.Context,
	cluster *v1alpha1.Cluster,
) (ClusterProvisioner, any, error) {
	if cluster == nil {
		return nil, nil, fmt.Errorf(
			"cluster configuration is required: %w",
			ErrUnsupportedDistribution,
		)
	}

	switch cluster.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return createKindProvisioner(
			cluster.Spec.DistributionConfig,
			cluster.Spec.Connection.Kubeconfig,
		)
	case v1alpha1.DistributionK3d:
		return createK3dProvisioner(cluster.Spec.DistributionConfig)
	default:
		return nil, "", fmt.Errorf("%w: %s", ErrUnsupportedDistribution, cluster.Spec.Distribution)
	}
}

func createKindProvisioner(
	distributionConfigPath string,
	kubeconfigPath string,
) (*kindprovisioner.KindClusterProvisioner, *v1alpha4.Cluster, error) {
	kindConfigMgr := kindconfigmanager.NewConfigManager(distributionConfigPath)

	err := kindConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load Kind configuration: %w", err)
	}

	provisioner, err := createKindProvisionerFromConfig(kindConfigMgr.GetConfig(), kubeconfigPath)
	if err != nil {
		return nil, nil, err
	}

	return provisioner, kindConfigMgr.GetConfig(), nil
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

func createK3dProvisioner(
	distributionConfigPath string,
) (*k3dprovisioner.K3dClusterProvisioner, *k3dv1alpha5.SimpleConfig, error) {
	k3dConfigMgr := k3dconfigmanager.NewConfigManager(distributionConfigPath)

	err := k3dConfigMgr.LoadConfig(nil)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to load K3d configuration: %w", err)
	}

	provisioner := createK3dProvisionerFromConfig(k3dConfigMgr.GetConfig())

	return provisioner, k3dConfigMgr.GetConfig(), nil
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
