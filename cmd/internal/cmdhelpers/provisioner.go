package cmdhelpers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	ksailio "github.com/devantler-tech/ksail-go/pkg/io"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	clusterprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster"
	k3dprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/k3d"
	kindprovisioner "github.com/devantler-tech/ksail-go/pkg/provisioner/cluster/kind"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// ErrUnsupportedDistribution is returned when an unsupported distribution is specified.
var ErrUnsupportedDistribution = errors.New("unsupported distribution")

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
	// Load Kind configuration
	kindConfig, err := loadKindConfig(cluster.Spec.DistributionConfig)
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
		kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
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
	// Load K3d configuration
	k3dConfig, err := loadK3dConfig(cluster.Spec.DistributionConfig)
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

// loadKindConfig loads and parses a Kind configuration file.
func loadKindConfig(configPath string) (*v1alpha4.Cluster, error) {
	// Find the config file
	resolvedPath, err := ksailio.FindFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find Kind config: %w", err)
	}

	// Read the file
	// #nosec G304 -- Path is validated by FindFile
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Kind config: %w", err)
	}

	// Unmarshal the YAML
	marshaller := yamlmarshaller.NewMarshaller[v1alpha4.Cluster]()

	var kindConfig v1alpha4.Cluster

	err = marshaller.Unmarshal(data, &kindConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Kind config: %w", err)
	}

	return &kindConfig, nil
}

// loadK3dConfig loads and parses a K3d configuration file.
func loadK3dConfig(configPath string) (*k3dv1alpha5.SimpleConfig, error) {
	// Find the config file
	resolvedPath, err := ksailio.FindFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to find K3d config: %w", err)
	}

	// Read the file
	// #nosec G304 -- Path is validated by FindFile
	data, err := os.ReadFile(resolvedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read K3d config: %w", err)
	}

	// Unmarshal the YAML
	marshaller := yamlmarshaller.NewMarshaller[k3dv1alpha5.SimpleConfig]()

	var k3dConfig k3dv1alpha5.SimpleConfig

	err = marshaller.Unmarshal(data, &k3dConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to parse K3d config: %w", err)
	}

	return &k3dConfig, nil
}
