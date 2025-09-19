// Package scaffolding provides utilities for scaffolding KSail project files and configuration.
package scaffolding

import (
	"errors"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	kindconfig "github.com/devantler-tech/ksail-go/pkg/config-manager/kind"
	eksgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// Scaffolder is responsible for generating KSail project files and configurations.
type Scaffolder struct {
	KSailConfig            v1alpha1.Cluster
	KSailYAMLGenerator     *yamlgenerator.YAMLGenerator[v1alpha1.Cluster]
	KindGenerator          *kindgenerator.KindGenerator
	K3dGenerator           *k3dgenerator.K3dGenerator
	EKSGenerator           *eksgenerator.EKSGenerator
	KustomizationGenerator *kustomizationgenerator.KustomizationGenerator
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
	if err := s.generateKSailConfig(output, force); err != nil {
		return err
	}

	if err := s.generateDistributionConfig(output, force); err != nil {
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

	return err
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
		return errors.New("talos-in-docker distribution is not yet implemented")
	default:
		return errors.New("provided distribution is unknown")
	}
}

// generateKindConfig generates the kind.yaml configuration file.
func (s *Scaffolder) generateKindConfig(output string, force bool) error {
	// Create a default Kind cluster configuration
	kindCluster := kindconfig.NewKindCluster(s.KSailConfig.Metadata.Name, "", "")
	// Add a minimal control plane node
	var node v1alpha4.Node
	node.Role = v1alpha4.ControlPlaneRole
	kindCluster.Nodes = append(kindCluster.Nodes, node)

	opts := yamlgenerator.Options{
		Output: output + "kind.yaml",
		Force:  force,
	}

	_, err := s.KindGenerator.Generate(kindCluster, opts)

	return err
}

// generateK3dConfig generates the k3d.yaml configuration file.
func (s *Scaffolder) generateK3dConfig(output string, force bool) error {
	opts := yamlgenerator.Options{
		Output: output + "k3d.yaml",
		Force:  force,
	}

	_, err := s.K3dGenerator.Generate(&s.KSailConfig, opts)

	return err
}

// generateEKSConfig generates the eks-config.yaml configuration file.
func (s *Scaffolder) generateEKSConfig(output string, force bool) error {
	eksCluster := createDefaultEKSConfig(s.KSailConfig.Metadata.Name)

	opts := yamlgenerator.Options{
		Output: output + "eks-config.yaml",
		Force:  force,
	}

	_, err := s.EKSGenerator.Generate(eksCluster, opts)

	return err
}

// generateKustomizationConfig generates the kustomization.yaml file.
func (s *Scaffolder) generateKustomizationConfig(output string, force bool) error {
	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory),
		Force:  force,
	}

	_, err := s.KustomizationGenerator.Generate(&s.KSailConfig, opts)

	return err
}

// createDefaultEKSConfig creates a minimal EKS cluster configuration for scaffolding.
func createDefaultEKSConfig(name string) *v1alpha5.ClusterConfig {
	minSize := 1
	maxSize := 3
	desiredCapacity := 2

	return &v1alpha5.ClusterConfig{
		TypeMeta: v1alpha5.ClusterConfigTypeMeta(),
		Metadata: &v1alpha5.ClusterMeta{
			Name:    name,
			Region:  "us-west-2",
			Version: "",
		},
		NodeGroups: []*v1alpha5.NodeGroup{
			{
				NodeGroupBase: &v1alpha5.NodeGroupBase{
					Name:         name + "-workers",
					InstanceType: "m5.large",
					ScalingConfig: &v1alpha5.ScalingConfig{
						MinSize:         &minSize,
						MaxSize:         &maxSize,
						DesiredCapacity: &desiredCapacity,
					},
				},
			},
		},
	}
}
