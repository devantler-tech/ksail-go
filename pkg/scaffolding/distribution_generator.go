// Package scaffolding provides utilities for scaffolding KSail project files and configuration.
package scaffolding

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	kindconfig "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	eksgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// DistributionGenerator generates distribution-specific configuration files.
type DistributionGenerator interface {
	Generate(cluster *v1alpha1.Cluster, opts yamlgenerator.Options) (string, error)
}

// KindDistributionGenerator generates Kind cluster configurations.
type KindDistributionGenerator struct {
	generator generator.Generator[*v1alpha4.Cluster, yamlgenerator.Options]
}

// NewKindDistributionGenerator creates a new Kind distribution generator.
func NewKindDistributionGenerator() *KindDistributionGenerator {
	return &KindDistributionGenerator{
		generator: kindgenerator.NewKindGenerator(),
	}
}

// Generate creates a Kind cluster configuration from a KSail cluster.
func (g *KindDistributionGenerator) Generate(
	cluster *v1alpha1.Cluster,
	opts yamlgenerator.Options,
) (string, error) {
	// Transform KSail cluster to Kind cluster
	kindCluster := kindconfig.NewKindCluster(cluster.Metadata.Name, "", "")
	// Add a minimal control plane node
	var node v1alpha4.Node

	node.Role = v1alpha4.ControlPlaneRole
	kindCluster.Nodes = append(kindCluster.Nodes, node)

	result, err := g.generator.Generate(kindCluster, opts)
	if err != nil {
		return "", fmt.Errorf("generate Kind configuration: %w", err)
	}

	return result, nil
}

// K3dDistributionGenerator generates K3d cluster configurations.
type K3dDistributionGenerator struct {
	generator generator.Generator[*v1alpha1.Cluster, yamlgenerator.Options]
}

// NewK3dDistributionGenerator creates a new K3d distribution generator.
func NewK3dDistributionGenerator() *K3dDistributionGenerator {
	return &K3dDistributionGenerator{
		generator: k3dgenerator.NewK3dGenerator(),
	}
}

// Generate creates a K3d cluster configuration from a KSail cluster.
func (g *K3dDistributionGenerator) Generate(
	cluster *v1alpha1.Cluster,
	opts yamlgenerator.Options,
) (string, error) {
	result, err := g.generator.Generate(cluster, opts)
	if err != nil {
		return "", fmt.Errorf("generate K3d configuration: %w", err)
	}

	return result, nil
}

// EKSDistributionGenerator generates EKS cluster configurations.
type EKSDistributionGenerator struct {
	generator generator.Generator[*v1alpha5.ClusterConfig, yamlgenerator.Options]
}

// NewEKSDistributionGenerator creates a new EKS distribution generator.
func NewEKSDistributionGenerator() *EKSDistributionGenerator {
	return &EKSDistributionGenerator{
		generator: eksgenerator.NewEKSGenerator(),
	}
}

// Generate creates an EKS cluster configuration from a KSail cluster.
func (g *EKSDistributionGenerator) Generate(
	cluster *v1alpha1.Cluster,
	opts yamlgenerator.Options,
) (string, error) {
	// Transform KSail cluster to EKS cluster
	eksCluster := createDefaultEKSConfig(cluster.Metadata.Name)

	result, err := g.generator.Generate(eksCluster, opts)
	if err != nil {
		return "", fmt.Errorf("generate EKS configuration: %w", err)
	}

	return result, nil
}

// createDefaultEKSConfig creates a minimal EKS cluster configuration for scaffolding.
func createDefaultEKSConfig(name string) *v1alpha5.ClusterConfig {
	minSize := 1
	maxSize := 3
	desiredCapacity := 2

	return &v1alpha5.ClusterConfig{
		TypeMeta:                v1alpha5.ClusterConfigTypeMeta(),
		Metadata:                createDefaultClusterMeta(name),
		KubernetesNetworkConfig: nil,
		AutoModeConfig:          nil,
		RemoteNetworkConfig:     nil,
		IAM:                     nil,
		IAMIdentityMappings:     nil,
		IdentityProviders:       nil,
		AccessConfig:            nil,
		VPC:                     nil,
		Addons:                  nil,
		AddonsConfig: v1alpha5.AddonsConfig{
			AutoApplyPodIdentityAssociations: false,
			DisableDefaultAddons:             false,
		},
		PrivateCluster:    nil,
		NodeGroups:        createDefaultNodeGroups(name, minSize, maxSize, desiredCapacity),
		ManagedNodeGroups: nil,
		FargateProfiles:   nil,
		AvailabilityZones: nil,
		LocalZones:        nil,
		CloudWatch:        nil,
		SecretsEncryption: nil,
		Status:            nil,
		GitOps:            nil,
		Karpenter:         nil,
		Outpost:           nil,
		ZonalShiftConfig:  nil,
	}
}

// createDefaultClusterMeta creates a default cluster metadata.
func createDefaultClusterMeta(name string) *v1alpha5.ClusterMeta {
	return &v1alpha5.ClusterMeta{
		Name:               name,
		Region:             "us-west-2",
		Version:            "",
		ForceUpdateVersion: nil,
		Tags:               nil,
		Annotations:        nil,
		AccountID:          "",
	}
}

// createDefaultNodeGroups creates default node groups for EKS.
func createDefaultNodeGroups(
	name string,
	minSize, maxSize, desiredCapacity int,
) []*v1alpha5.NodeGroup {
	return []*v1alpha5.NodeGroup{
		{
			NodeGroupBase: createDefaultNodeGroupBase(name, minSize, maxSize, desiredCapacity),
		},
	}
}

// createDefaultNodeGroupBase creates a default node group base with all required fields.
func createDefaultNodeGroupBase(
	name string,
	minSize, maxSize, desiredCapacity int,
) *v1alpha5.NodeGroupBase {
	return &v1alpha5.NodeGroupBase{
		Name:                      name + "-workers",
		AMIFamily:                 "",
		InstanceType:              "m5.large",
		AvailabilityZones:         nil,
		Subnets:                   nil,
		InstancePrefix:            "",
		InstanceName:              "",
		VolumeSize:                nil,
		SSH:                       nil,
		Labels:                    nil,
		PrivateNetworking:         false,
		Tags:                      nil,
		IAM:                       nil,
		AMI:                       "",
		SecurityGroups:            nil,
		MaxPodsPerNode:            0,
		ASGSuspendProcesses:       nil,
		EBSOptimized:              nil,
		VolumeType:                nil,
		VolumeName:                nil,
		VolumeEncrypted:           nil,
		VolumeKmsKeyID:            nil,
		VolumeIOPS:                nil,
		VolumeThroughput:          nil,
		AdditionalVolumes:         nil,
		PreBootstrapCommands:      nil,
		OverrideBootstrapCommand:  nil,
		PropagateASGTags:          nil,
		DisableIMDSv1:             nil,
		DisablePodIMDS:            nil,
		Placement:                 nil,
		EFAEnabled:                nil,
		InstanceSelector:          nil,
		AdditionalEncryptedVolume: "",
		Bottlerocket:              nil,
		EnableDetailedMonitoring:  nil,
		CapacityReservation:       nil,
		InstanceMarketOptions:     nil,
		OutpostARN:                "",
		ScalingConfig: &v1alpha5.ScalingConfig{
			MinSize:         &minSize,
			MaxSize:         &maxSize,
			DesiredCapacity: &desiredCapacity,
		},
	}
}

// NewDistributionGenerator creates the appropriate distribution generator based on the distribution type.
//
//nolint:ireturn // Factory pattern requires returning interface
func NewDistributionGenerator(
	distribution v1alpha1.Distribution,
) (DistributionGenerator, error) {
	switch distribution {
	case v1alpha1.DistributionKind:
		return NewKindDistributionGenerator(), nil
	case v1alpha1.DistributionK3d:
		return NewK3dDistributionGenerator(), nil
	case v1alpha1.DistributionEKS:
		return NewEKSDistributionGenerator(), nil
	case v1alpha1.DistributionTind:
		return nil, ErrTindNotImplemented
	default:
		return nil, ErrUnknownDistribution
	}
}
