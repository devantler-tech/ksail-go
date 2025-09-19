// Package scaffolding provides utilities for scaffolding KSail project files and configuration.
package scaffolding

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	eksgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Error definitions for distribution handling.
var (
	ErrTindNotImplemented      = errors.New("talos-in-docker distribution is not yet implemented")
	ErrUnknownDistribution     = errors.New("provided distribution is unknown")
	ErrKSailConfigGeneration   = errors.New("failed to generate KSail configuration")
	ErrKindConfigGeneration    = errors.New("failed to generate Kind configuration")
	ErrK3dConfigGeneration     = errors.New("failed to generate K3d configuration")
	ErrEKSConfigGeneration     = errors.New("failed to generate EKS configuration")
	ErrKustomizationGeneration = errors.New("failed to generate kustomization configuration")
)

// Scaffolder is responsible for generating KSail project files and configurations.
type Scaffolder struct {
	KSailConfig            v1alpha1.Cluster
	KSailYAMLGenerator     generator.Generator[v1alpha1.Cluster, yamlgenerator.Options]
	KindGenerator          generator.Generator[*v1alpha4.Cluster, yamlgenerator.Options]
	K3dGenerator           generator.Generator[*v1alpha1.Cluster, yamlgenerator.Options]
	EKSGenerator           generator.Generator[*v1alpha5.ClusterConfig, yamlgenerator.Options]
	KustomizationGenerator generator.Generator[*v1alpha1.Cluster, yamlgenerator.Options]
}

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
func NewScaffolder(cfg v1alpha1.Cluster) *Scaffolder {
	ksailGen := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()
	kindGen := kindgenerator.NewKindGenerator()
	k3dGen := k3dgenerator.NewK3dGenerator()
	eksGen := eksgenerator.NewEKSGenerator()
	kustGen := kustomizationgenerator.NewKustomizationGenerator(&cfg)

	return &Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGen,
		KindGenerator:          kindGen,
		K3dGenerator:           k3dGen,
		EKSGenerator:           eksGen,
		KustomizationGenerator: kustGen,
	}
}

// Scaffold generates project files and configurations.
func (s *Scaffolder) Scaffold(output string, force bool) error {
	err := s.generateKSailConfig(output, force)
	if err != nil {
		return err
	}

	err = s.generateDistributionConfig(output, force)
	if err != nil {
		return err
	}

	return s.generateKustomizationConfig(output, force)
}

// generateKSailConfig generates the ksail.yaml configuration file.
func (s *Scaffolder) generateKSailConfig(output string, force bool) error {
	opts := yamlgenerator.Options{
		Output: output + "ksail.yaml",
		Force:  force,
	}

	_, err := s.KSailYAMLGenerator.Generate(s.KSailConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKSailConfigGeneration, err)
	}

	return nil
}

// generateDistributionConfig generates the distribution-specific configuration file.
func (s *Scaffolder) generateDistributionConfig(output string, force bool) error {
	switch s.KSailConfig.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return s.generateKindConfig(output, force)
	case v1alpha1.DistributionK3d:
		return s.generateK3dConfig(output, force)
	case v1alpha1.DistributionEKS:
		return s.generateEKSConfig(output, force)
	case v1alpha1.DistributionTind:
		return ErrTindNotImplemented
	default:
		return ErrUnknownDistribution
	}
}

// generateKindConfig generates the kind.yaml configuration file.
func (s *Scaffolder) generateKindConfig(output string, force bool) error {
	// Create a minimal Kind cluster configuration
	kindCluster := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			APIVersion: "kind.x-k8s.io/v1alpha4",
			Kind:       "Cluster",
		},
		Name: s.KSailConfig.Metadata.Name,
	}

	opts := yamlgenerator.Options{
		Output: output + "kind.yaml",
		Force:  force,
	}

	_, err := s.KindGenerator.Generate(kindCluster, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKindConfigGeneration, err)
	}

	return nil
}

// generateK3dConfig generates the k3d.yaml configuration file.
func (s *Scaffolder) generateK3dConfig(output string, force bool) error {
	opts := yamlgenerator.Options{
		Output: output + "k3d.yaml",
		Force:  force,
	}

	_, err := s.K3dGenerator.Generate(&s.KSailConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrK3dConfigGeneration, err)
	}

	return nil
}

// generateEKSConfig generates the eks-config.yaml configuration file.
func (s *Scaffolder) generateEKSConfig(output string, force bool) error {
	eksCluster := createDefaultEKSConfig(s.KSailConfig.Metadata.Name)

	opts := yamlgenerator.Options{
		Output: output + "eks-config.yaml",
		Force:  force,
	}

	_, err := s.EKSGenerator.Generate(eksCluster, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEKSConfigGeneration, err)
	}

	return nil
}

// generateKustomizationConfig generates the kustomization.yaml file.
func (s *Scaffolder) generateKustomizationConfig(output string, force bool) error {
	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory),
		Force:  force,
	}

	_, err := s.KustomizationGenerator.Generate(&s.KSailConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKustomizationGeneration, err)
	}

	return nil
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
