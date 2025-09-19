// Package scaffolding provides utilities for scaffolding KSail project files and configuration.
package scaffolding

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
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
	KustomizationGenerator generator.Generator[*v1alpha1.Cluster, yamlgenerator.Options]
}

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
func NewScaffolder(cfg v1alpha1.Cluster) *Scaffolder {
	ksailGen := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()
	kustGen := kustomizationgenerator.NewKustomizationGenerator(&cfg)

	return &Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGen,
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
	// Create a minimal Kind cluster configuration with explicit minimal YAML output
	yamlContent := fmt.Sprintf(`apiVersion: kind.x-k8s.io/v1alpha4
kind: Cluster
name: %s
`, s.KSailConfig.Metadata.Name)

	opts := yamlgenerator.Options{
		Output: output + "kind.yaml",
		Force:  force,
	}

	// Write minimal YAML directly instead of using generator to avoid defaults
	if opts.Output != "" {
		result, err := io.TryWriteFile(yamlContent, opts.Output, opts.Force)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrKindConfigGeneration, err)
		}

		_ = result
	}

	return nil
}

// generateK3dConfig generates the k3d.yaml configuration file.
func (s *Scaffolder) generateK3dConfig(output string, force bool) error {
	// Create minimal K3d YAML content without extra fields
	yamlContent := fmt.Sprintf("apiVersion: k3d.io/v1alpha5\nkind: Simple\nmetadata:\n  name: %s\n", s.KSailConfig.Metadata.Name)

	opts := yamlgenerator.Options{
		Output: output + "k3d.yaml",
		Force:  force,
	}

	// Write minimal YAML directly
	if opts.Output != "" {
		result, err := io.TryWriteFile(yamlContent, opts.Output, opts.Force)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrK3dConfigGeneration, err)
		}

		_ = result
	}

	return nil
}

// generateEKSConfig generates the eks-config.yaml configuration file.
func (s *Scaffolder) generateEKSConfig(output string, force bool) error {
	// Create minimal EKS YAML content without extra fields
	name := s.KSailConfig.Metadata.Name
	yamlContent := fmt.Sprintf(`apiVersion: eksctl.io/v1alpha5
kind: ClusterConfig
metadata:
  name: %s
  region: us-west-2
nodeGroups:
- desiredCapacity: 2
  instanceType: m5.large
  maxSize: 3
  minSize: 1
  name: %s-workers
`, name, name)

	opts := yamlgenerator.Options{
		Output: output + "eks-config.yaml",
		Force:  force,
	}

	// Write minimal YAML directly
	if opts.Output != "" {
		result, err := io.TryWriteFile(yamlContent, opts.Output, opts.Force)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrEKSConfigGeneration, err)
		}

		_ = result
	}

	return nil
}

// generateKustomizationConfig generates the kustomization.yaml file.
func (s *Scaffolder) generateKustomizationConfig(output string, force bool) error {
	// Create minimal Kustomization YAML content with empty resources array
	yamlContent := "apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources: []\n"

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory, "kustomization.yaml"),
		Force:  force,
	}

	// Write minimal YAML directly
	if opts.Output != "" {
		result, err := io.TryWriteFile(yamlContent, opts.Output, opts.Force)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrKustomizationGeneration, err)
		}

		_ = result
	}

	return nil
}



