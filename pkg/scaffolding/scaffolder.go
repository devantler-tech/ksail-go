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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		Name:                 s.KSailConfig.Metadata.Name,
		Nodes:                []v1alpha4.Node{},
		Networking:           v1alpha4.Networking{},
		FeatureGates:         map[string]bool{},
		RuntimeConfig:        map[string]string{},
		KubeadmConfigPatches: []string{},
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
	// Create EKS cluster configuration with required fields
	eksConfig := &v1alpha5.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eksctl.io/v1alpha5",
			Kind:       "ClusterConfig",
		},
		Metadata: &v1alpha5.ClusterMeta{
			Name:   s.KSailConfig.Metadata.Name,
			Region: "us-west-2",
		},
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name:         s.KSailConfig.Metadata.Name + "-workers",
					InstanceType: "m5.large",
					ScalingConfig: &v1alpha5.ScalingConfig{
						DesiredCapacity: &[]int{2}[0],
						MinSize:         &[]int{1}[0],
						MaxSize:         &[]int{3}[0],
					},
					VolumeSize:        &[]int{0}[0],
					VolumeType:        &[]string{""}[0],
					VolumeEncrypted:   &[]bool{false}[0],
					Tags:              map[string]string{},
					PrivateNetworking: false,
				},
			},
		},
		ManagedNodeGroups: []*v1alpha5.ManagedNodeGroup{},
		FargateProfiles:   []*v1alpha5.FargateProfile{},
		AvailabilityZones: []string{},
		CloudWatch: &v1alpha5.ClusterCloudWatch{
			ClusterLogging: &v1alpha5.ClusterCloudWatchLogging{
				EnableTypes:        []string{},
				LogRetentionInDays: 0,
			},
		},
		SecretsEncryption: &v1alpha5.SecretsEncryption{
			KeyARN: "",
		},
	}

	opts := yamlgenerator.Options{
		Output: output + "eks-config.yaml",
		Force:  force,
	}

	_, err := s.EKSGenerator.Generate(eksConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrEKSConfigGeneration, err)
	}

	return nil
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



