// Package scaffolder provides utilities for scaffolding KSail project files and configuration.
package scaffolder

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/cmd/ui/notify"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	eksgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	eksv1alpha5 "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	ktypes "sigs.k8s.io/kustomize/api/types"
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

// Distribution config file constants.
const (
	KindConfigFile = "kind.yaml"
	K3dConfigFile  = "k3d.yaml"
	EKSConfigFile  = "eks.yaml"
	TindConfigFile = "tind.yaml"
)

// getExpectedContextName calculates the expected context name for a distribution using default cluster names.
// This is used during scaffolding to ensure consistent context patterns between KSail config and distribution configs.
// Returns empty string for EKS (context validation is skipped) and unsupported distributions.
func getExpectedContextName(distribution v1alpha1.Distribution) string {
	var distributionName string

	switch distribution {
	case v1alpha1.DistributionKind:
		distributionName = "kind" // Default Kind cluster name (matches generateKindConfig)

		return "kind-" + distributionName
	case v1alpha1.DistributionK3d:
		distributionName = "k3s-default" // Default K3d cluster name (matches createK3dConfig)

		return "k3d-" + distributionName
	case v1alpha1.DistributionEKS:
		// EKS context validation is skipped, return empty
		return ""
	case v1alpha1.DistributionTind:
		// Tind is not yet implemented
		return ""
	default:
		return ""
	}
}

// getExpectedDistributionConfigName returns the expected distribution config filename for a distribution.
// This is used during scaffolding to set the correct config file name that matches the generated files.
func getExpectedDistributionConfigName(distribution v1alpha1.Distribution) string {
	switch distribution {
	case v1alpha1.DistributionKind:
		return KindConfigFile
	case v1alpha1.DistributionK3d:
		return K3dConfigFile
	case v1alpha1.DistributionEKS:
		return EKSConfigFile
	case v1alpha1.DistributionTind:
		return TindConfigFile
	default:
		return KindConfigFile // fallback default
	}
}

// Scaffolder is responsible for generating KSail project files and configurations.
type Scaffolder struct {
	KSailConfig            v1alpha1.Cluster
	KSailYAMLGenerator     generator.Generator[v1alpha1.Cluster, yamlgenerator.Options]
	KindGenerator          generator.Generator[*v1alpha4.Cluster, yamlgenerator.Options]
	K3dGenerator           generator.Generator[*k3dv1alpha5.SimpleConfig, yamlgenerator.Options]
	EKSGenerator           generator.Generator[*eksv1alpha5.ClusterConfig, yamlgenerator.Options]
	KustomizationGenerator generator.Generator[*ktypes.Kustomization, yamlgenerator.Options]
}

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
func NewScaffolder(cfg v1alpha1.Cluster) *Scaffolder {
	ksailGenerator := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()
	kindGenerator := kindgenerator.NewKindGenerator()
	k3dGenerator := k3dgenerator.NewK3dGenerator()
	eksGenerator := eksgenerator.NewEKSGenerator()
	kustomizationGenerator := kustomizationgenerator.NewKustomizationGenerator()

	return &Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGenerator,
		KindGenerator:          kindGenerator,
		K3dGenerator:           k3dGenerator,
		EKSGenerator:           eksGenerator,
		KustomizationGenerator: kustomizationGenerator,
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

// applyKSailConfigDefaults applies distribution-specific defaults to the KSail configuration.
// This ensures the generated ksail.yaml has consistent context and distributionConfig values
// that match the distribution-specific configuration files being generated.
func (s *Scaffolder) applyKSailConfigDefaults() v1alpha1.Cluster {
	config := s.KSailConfig

	// Set the expected context if it's empty, based on the distribution and default cluster names
	if config.Spec.Connection.Context == "" {
		expectedContext := getExpectedContextName(config.Spec.Distribution)
		if expectedContext != "" {
			config.Spec.Connection.Context = expectedContext
		}
	}

	// Set the expected distribution config filename if it's empty or set to default
	if config.Spec.DistributionConfig == "" || config.Spec.DistributionConfig == KindConfigFile {
		expectedConfigName := getExpectedDistributionConfigName(config.Spec.Distribution)
		config.Spec.DistributionConfig = expectedConfigName
	}

	return config
}

// generateKSailConfig generates the ksail.yaml configuration file.
func (s *Scaffolder) generateKSailConfig(output string, force bool) error {
	// Apply distribution-specific defaults to ensure consistency with generated files
	config := s.applyKSailConfigDefaults()

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, "ksail.yaml"),
		Force:  force,
	}

	if _, err := os.Stat(opts.Output); err == nil && !force {
		notify.Warnln(os.Stdout, "skipped 'ksail.yaml', file exists use --force to overwrite")
		return nil
	}

	_, err := s.KSailYAMLGenerator.Generate(config, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKSailConfigGeneration, err)
	}
	if force {
		notify.Successln(os.Stdout, "overwrote 'ksail.yaml'")
	} else {
		notify.Successln(os.Stdout, "created 'ksail.yaml'")
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
	// Create Kind cluster configuration with standard KSail name
	kindConfig := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			APIVersion: "kind.x-k8s.io/v1alpha4",
			Kind:       "Cluster",
		},
		Name: "kind",
	}

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, KindConfigFile),
		Force:  force,
	}

	if _, err := os.Stat(opts.Output); err == nil && !force {
		notify.Warnln(os.Stdout, "skipped 'kind.yaml', file exists use --force to overwrite")
		return nil
	}

	_, err := s.KindGenerator.Generate(kindConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKindConfigGeneration, err)
	}
	if force {
		notify.Successln(os.Stdout, "overwrote 'kind.yaml'")
	} else {
		notify.Successln(os.Stdout, "created 'kind.yaml'")
	}

	return nil
}

// generateK3dConfig generates the k3d.yaml configuration file.
func (s *Scaffolder) generateK3dConfig(output string, force bool) error {
	k3dConfig := s.createK3dConfig()

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, "k3d.yaml"),
		Force:  force,
	}

	if _, err := os.Stat(opts.Output); err == nil && !force {
		notify.Warnln(os.Stdout, "skipped 'k3d.yaml', file exists use --force to overwrite")
		return nil
	}

	_, err := s.K3dGenerator.Generate(&k3dConfig, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrK3dConfigGeneration, err)
	}
	if force {
		notify.Successln(os.Stdout, "overwrote 'k3d.yaml'")
	} else {
		notify.Successln(os.Stdout, "created 'k3d.yaml'")
	}

	return nil
}

// generateEKSConfig generates the eks.yaml configuration file.
func (s *Scaffolder) generateEKSConfig(output string, force bool) error {
	eksConfig := s.createEKSConfig()

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, "eks.yaml"),
		Force:  force,
	}

	if _, err := os.Stat(opts.Output); err == nil && !force {
		notify.Warnln(os.Stdout, "skipped 'eks.yaml', file exists use --force to overwrite")
		return nil
	}

	_, err := s.EKSGenerator.Generate(eksConfig, opts)
	if err != nil {
		return fmt.Errorf("generate EKS config: %w", err)
	}
	if force {
		notify.Successln(os.Stdout, "overwrote 'eks.yaml'")
	} else {
		notify.Successln(os.Stdout, "created 'eks.yaml'")
	}

	return nil
}

func (s *Scaffolder) createK3dConfig() k3dv1alpha5.SimpleConfig {
	config := k3dv1alpha5.SimpleConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: "k3d.io/v1alpha5",
			Kind:       "Simple",
		},
		ObjectMeta: types.ObjectMeta{
			Name: "k3s-default",
		},
	}

	return config
}

func (s *Scaffolder) createEKSConfig() *eksv1alpha5.ClusterConfig {
	return &eksv1alpha5.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eksctl.io/v1alpha5",
			Kind:       "ClusterConfig",
		},
		Metadata: &eksv1alpha5.ClusterMeta{
			Name:   "eks-default",
			Region: "eu-north-1",
		},
	}
}

// generateKustomizationConfig generates the kustomization.yaml file.
func (s *Scaffolder) generateKustomizationConfig(output string, force bool) error {
	kustomization := ktypes.Kustomization{}

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory, "kustomization.yaml"),
		Force:  force,
	}

	if _, err := os.Stat(opts.Output); err == nil && !force {
		notify.Warnln(
			os.Stdout,
			"skipped 'k8s/kustomization.yaml', file exists use --force to overwrite",
		)
		return nil
	}

	_, err := s.KustomizationGenerator.Generate(&kustomization, opts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrKustomizationGeneration, err)
	}
	if force {
		notify.Successln(os.Stdout, "overwrote 'k8s/kustomization.yaml'")
	} else {
		notify.Successln(os.Stdout, "created 'k8s/kustomization.yaml'")
	}

	return nil
}
