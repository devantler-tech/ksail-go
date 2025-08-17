// Package scaffolder provides utilities for scaffolding KSail project files and configuration.
package scaffolder

import (
	"errors"
	"path/filepath"

	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
	gen "github.com/devantler-tech/ksail-go/pkg/generator"
)

// Scaffolder is responsible for generating KSail project files and configurations.
type Scaffolder struct {
	KSailConfig            ksailcluster.Cluster
	KSailYAMLGenerator     *gen.YAMLGenerator[ksailcluster.Cluster]
	KindGenerator          *gen.KindGenerator
	K3dGenerator           *gen.K3dGenerator
	KustomizationGenerator *gen.KustomizationGenerator
}

// Scaffold generates project files and configurations.
func (s *Scaffolder) Scaffold(output string, force bool) error {
	// generate ksail.yaml file
	_, err := s.KSailYAMLGenerator.Generate(s.KSailConfig, gen.Options{Output: output + "ksail.yaml", Force: force})
	if err != nil {
		return err
	}

	// generate distribution config file
	switch s.KSailConfig.Spec.Distribution {
	case ksailcluster.DistributionKind:
		if _, err := s.KindGenerator.Generate(gen.Options{Output: output + "kind.yaml", Force: force}); err != nil {
			return err
		}
	case ksailcluster.DistributionK3d:
		if _, err := s.K3dGenerator.Generate(gen.Options{Output: output + "k3d.yaml", Force: force}); err != nil {
			return err
		}
	case ksailcluster.DistributionTind:
		return errors.New("talos-in-docker distribution is not yet implemented")
	default:
		return errors.New("provided distribution is unknown")
	}

	if _, err := s.KustomizationGenerator.Generate(gen.Options{Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory), Force: force}); err != nil {
		return err
	}

	return nil
}

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
func NewScaffolder(cfg ksailcluster.Cluster) *Scaffolder {
	ksailGen := gen.NewYAMLGenerator[ksailcluster.Cluster]()
	kindGen := gen.NewKindGenerator(&cfg)
	k3dGen := gen.NewK3dGenerator(&cfg)
	kustGen := gen.NewKustomizationGenerator(&cfg)

	return &Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGen,
		KindGenerator:          kindGen,
		K3dGenerator:           k3dGen,
		KustomizationGenerator: kustGen,
	}
}

// --- internals ---
