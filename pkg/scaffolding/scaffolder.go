// Package scaffolding provides utilities for scaffolding KSail project files and configuration.
package scaffolding

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
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
	DistributionGenerator  DistributionGenerator
	KustomizationGenerator generator.Generator[*v1alpha1.Cluster, yamlgenerator.Options]
}

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
func NewScaffolder(cfg v1alpha1.Cluster) (*Scaffolder, error) {
	ksailGen := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()
	kustGen := kustomizationgenerator.NewKustomizationGenerator(&cfg)

	distGen, err := NewDistributionGenerator(cfg.Spec.Distribution)
	if err != nil {
		return nil, err
	}

	return &Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGen,
		DistributionGenerator:  distGen,
		KustomizationGenerator: kustGen,
	}, nil
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
	var filename string
	//nolint:exhaustive // Tind is handled in NewDistributionGenerator
	switch s.KSailConfig.Spec.Distribution {
	case v1alpha1.DistributionKind:
		filename = "kind.yaml"
	case v1alpha1.DistributionK3d:
		filename = "k3d.yaml"
	case v1alpha1.DistributionEKS:
		filename = "eks-config.yaml"
	default:
		// This should not happen since NewScaffolder validates the distribution
		return ErrUnknownDistribution
	}

	opts := yamlgenerator.Options{
		Output: output + filename,
		Force:  force,
	}

	_, err := s.DistributionGenerator.Generate(&s.KSailConfig, opts)
	if err != nil {
		return fmt.Errorf(
			"failed to generate %s configuration: %w",
			s.KSailConfig.Spec.Distribution,
			err,
		)
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
