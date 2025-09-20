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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// Create a copy of the config and filter out default values
	config := s.KSailConfig

	// Filter out default values to keep output minimal
	if config.Spec.SourceDirectory == "k8s" {
		config.Spec.SourceDirectory = ""
	}

	if config.Spec.Distribution == v1alpha1.DistributionKind {
		config.Spec.Distribution = ""
	}

	if config.Spec.DistributionConfig == "kind.yaml" {
		config.Spec.DistributionConfig = ""
	}

	opts := yamlgenerator.Options{
		Output: output + "ksail.yaml",
		Force:  force,
	}

	_, err := s.KSailYAMLGenerator.Generate(config, opts)
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
	// Create minimal Kind cluster configuration
	kindConfig := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			APIVersion: "kind.x-k8s.io/v1alpha4",
			Kind:       "Cluster",
		},
		Name: s.KSailConfig.Metadata.Name,
		Nodes: []v1alpha4.Node{},
		Networking: v1alpha4.Networking{
			IPFamily:            "",
			APIServerPort:       0,
			APIServerAddress:    "",
			PodSubnet:           "",
			ServiceSubnet:       "",
			DisableDefaultCNI:   false,
			KubeProxyMode:       "",
			DNSSearch:           nil,
		},
		FeatureGates:                     map[string]bool{},
		RuntimeConfig:                    map[string]string{},
		KubeadmConfigPatches:             []string{},
		KubeadmConfigPatchesJSON6902:     nil,
		ContainerdConfigPatches:          []string{},
		ContainerdConfigPatchesJSON6902:  nil,
	}

	opts := yamlgenerator.Options{
		Output: output + "kind.yaml",
		Force:  force,
	}

	_, err := s.KindGenerator.Generate(kindConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKindConfigGeneration, err)
	}

	return nil
}

// generateK3dConfig generates the k3d.yaml configuration file.
func (s *Scaffolder) generateK3dConfig(output string, force bool) error {
	// Create minimal K3d configuration
	k3dConfig := &s.KSailConfig

	opts := yamlgenerator.Options{
		Output: output + "k3d.yaml",
		Force:  force,
	}

	_, err := s.K3dGenerator.Generate(k3dConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrK3dConfigGeneration, err)
	}

	return nil
}

// generateEKSConfig generates the eks-config.yaml configuration file.
func (s *Scaffolder) generateEKSConfig(output string, force bool) error {
	eksConfig := s.createMinimalEKSConfig()
	
	eksGen := eksgenerator.NewEKSGenerator()
	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.DistributionConfig),
		Force:  force,
	}

	_, err := eksGen.Generate(eksConfig, opts)
	if err != nil {
		return fmt.Errorf("generate EKS config: %w", err)
	}
	
	return nil
}

func (s *Scaffolder) createMinimalEKSConfig() *v1alpha5.ClusterConfig {
	return &v1alpha5.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eksctl.io/v1alpha5",
			Kind:       "ClusterConfig",
		},
		Metadata: s.createEKSMetadata(),
		NodeGroups: []*v1alpha5.NodeGroup{
			s.createEKSNodeGroup(),
		},
		ManagedNodeGroups:       nil,
		FargateProfiles:         nil,
		AvailabilityZones:       nil,
		LocalZones:              nil,
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
		PrivateCluster:   nil,
		CloudWatch:       nil,
		SecretsEncryption: nil,
		Status:           nil,
		GitOps:           nil,
		Karpenter:        nil,
		Outpost:          nil,
		ZonalShiftConfig: nil,
	}
}

func (s *Scaffolder) createEKSMetadata() *v1alpha5.ClusterMeta {
	return &v1alpha5.ClusterMeta{
		Name:              s.KSailConfig.Metadata.Name,
		Region:            "eu-north-1",
		Version:           "",        // string
		ForceUpdateVersion: nil,
		Tags:              nil,
		Annotations:       nil,
		AccountID:         "",        // string
	}
}

func (s *Scaffolder) createEKSNodeGroup() *v1alpha5.NodeGroup {
	return &v1alpha5.NodeGroup{
		NodeGroupBase: s.createEKSNodeGroupBase(),
		InstancesDistribution:    nil,
		ASGMetricsCollection:     nil,
		CPUCredits:               nil,
		ClassicLoadBalancerNames: nil,
		TargetGroupARNs:          nil,
		Taints:                   nil,
		UpdateConfig:             nil,
		ClusterDNS:               "", // string zero value
		KubeletExtraConfig:       nil,
		ContainerRuntime:         nil,
		MaxInstanceLifetime:      nil,
		LocalZones:               nil,
		EnclaveEnabled:           nil,
	}
}

func (s *Scaffolder) createEKSNodeGroupBase() *v1alpha5.NodeGroupBase {
	return &v1alpha5.NodeGroupBase{
		Name:                        "ng-1",
		AMIFamily:                   "", // string zero value
		InstanceType:                "m5.large",
		AvailabilityZones:           nil,
		Subnets:                     nil,
		InstancePrefix:              "", // string zero value
		InstanceName:                "", // string zero value
		ScalingConfig: &v1alpha5.ScalingConfig{
			DesiredCapacity: &[]int{1}[0],
			MinSize:         nil,
			MaxSize:         nil,
		},
		VolumeSize:                  nil,
		VolumeType:                  nil,
		VolumeEncrypted:             nil,
		VolumeKmsKeyID:              nil, // *string
		VolumeIOPS:                  nil,
		VolumeThroughput:            nil,
		VolumeName:                  nil, // *string
		AdditionalVolumes:           nil,
		SSH:                         nil,
		Labels:                      nil,
		IAM:                         nil,
		AMI:                         "", // string
		SecurityGroups:              nil,
		MaxPodsPerNode:              0,  // int zero value
		ASGSuspendProcesses:         nil,
		EBSOptimized:                nil,
		PreBootstrapCommands:        nil,
		OverrideBootstrapCommand:    nil, // *string
		Tags:                        nil,
		PropagateASGTags:            nil,
		DisableIMDSv1:               nil,
		DisablePodIMDS:              nil,
		Placement:                   nil,
		EFAEnabled:                  nil,
		InstanceSelector:            nil,
		AdditionalEncryptedVolume:   "", // string
		Bottlerocket:                nil,
		EnableDetailedMonitoring:    nil,
		CapacityReservation:         nil,
		InstanceMarketOptions:       nil,
		OutpostARN:                  "", // string
		PrivateNetworking:           false,
	}
}

// generateKustomizationConfig generates the kustomization.yaml file.
func (s *Scaffolder) generateKustomizationConfig(output string, force bool) error {
	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory, "kustomization.yaml"),
		Force:  force,
	}

	_, err := s.KustomizationGenerator.Generate(&s.KSailConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKustomizationGeneration, err)
	}

	return nil
}



