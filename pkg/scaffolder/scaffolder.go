// Package scaffolder provides utilities for scaffolding KSail project files and configuration.
package scaffolder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

// Error definitions for distribution handling.
var (
	ErrUnknownDistribution     = errors.New("provided distribution is unknown")
	ErrKSailConfigGeneration   = errors.New("failed to generate KSail configuration")
	ErrKindConfigGeneration    = errors.New("failed to generate Kind configuration")
	ErrK3dConfigGeneration     = errors.New("failed to generate K3d configuration")
	ErrKustomizationGeneration = errors.New("failed to generate kustomization configuration")
)

// Distribution config file constants.
const (
	KindConfigFile = "kind.yaml"
	K3dConfigFile  = "k3d.yaml"
)

// getExpectedContextName calculates the expected context name for a distribution using default cluster names.
// This is used during scaffolding to ensure consistent context patterns between KSail config and distribution configs.
// Returns empty string for unsupported distributions.
func getExpectedContextName(distribution v1alpha1.Distribution) string {
	var distributionName string

	switch distribution {
	case v1alpha1.DistributionKind:
		distributionName = "kind" // Default Kind cluster name (matches generateKindConfig)

		return "kind-" + distributionName
	case v1alpha1.DistributionK3d:
		distributionName = "k3s-default" // Default K3d cluster name (matches createK3dConfig)

		return "k3d-" + distributionName
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
	KustomizationGenerator generator.Generator[*ktypes.Kustomization, yamlgenerator.Options]
	Writer                 io.Writer
}

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
func NewScaffolder(cfg v1alpha1.Cluster, writer io.Writer) *Scaffolder {
	ksailGenerator := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()
	kindGenerator := kindgenerator.NewKindGenerator()
	k3dGenerator := k3dgenerator.NewK3dGenerator()
	kustomizationGenerator := kustomizationgenerator.NewKustomizationGenerator()

	return &Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGenerator,
		KindGenerator:          kindGenerator,
		K3dGenerator:           k3dGenerator,
		KustomizationGenerator: kustomizationGenerator,
		Writer:                 writer,
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

// checkFileExistsAndSkip checks if a file exists and should be skipped based on force flag.
// Returns true if the file should be skipped (exists and force=false), false otherwise.
// Outputs appropriate warning message if skipping.
func (s *Scaffolder) checkFileExistsAndSkip(
	filePath string,
	fileName string,
	force bool,
) (bool, bool, time.Time) {
	info, statErr := os.Stat(filePath)
	if statErr == nil {
		if !force {
			notify.WriteMessage(notify.Message{
				Type:    notify.WarningType,
				Content: "skipped '%s', file exists use --force to overwrite",
				Args:    []any{fileName},
				Writer:  s.Writer,
			})

			return true, true, info.ModTime()
		}

		return false, true, info.ModTime()
	}

	if statErr != nil && !errors.Is(statErr, os.ErrNotExist) {
		return false, false, time.Time{}
	}

	return false, false, time.Time{}
}

// GenerationParams groups parameters for generateWithFileHandling.
type GenerationParams[T any] struct {
	Gen         generator.Generator[T, yamlgenerator.Options]
	Model       T
	Opts        yamlgenerator.Options
	DisplayName string
	Force       bool
	WrapErr     func(error) error
}

// generateWithFileHandling wraps template generation with common file existence checks and notifications.

func generateWithFileHandling[T any](
	scaffolder *Scaffolder,
	params GenerationParams[T],
) error {
	skip, existed, previousModTime := scaffolder.checkFileExistsAndSkip(
		params.Opts.Output,
		params.DisplayName,
		params.Force,
	)

	if skip {
		return nil
	}

	_, err := params.Gen.Generate(params.Model, params.Opts)
	if err != nil {
		if params.WrapErr != nil {
			return params.WrapErr(err)
		}

		return fmt.Errorf("failed to generate %s: %w", params.DisplayName, err)
	}

	if params.Force && existed {
		err := ensureOverwriteModTime(params.Opts.Output, previousModTime)
		if err != nil {
			return fmt.Errorf("failed to update mod time for %s: %w", params.DisplayName, err)
		}
	}

	scaffolder.notifyFileAction(params.DisplayName, existed)

	return nil
}

func ensureOverwriteModTime(path string, previous time.Time) error {
	if path == "" {
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}

	current := info.ModTime()
	if previous.IsZero() || current.After(previous) {
		return nil
	}

	// Ensure the new mod time is strictly greater than the previous timestamp.
	newModTime := previous.Add(time.Millisecond)

	now := time.Now()
	if now.After(newModTime) {
		newModTime = now
	}

	err = os.Chtimes(path, newModTime, newModTime)
	if err != nil {
		return fmt.Errorf("failed to update mod time for %s: %w", path, err)
	}

	return nil
}

func (s *Scaffolder) notifyFileAction(displayName string, overwritten bool) {
	action := "created"
	if overwritten {
		action = "overwrote"
	}

	notify.WriteMessage(notify.Message{
		Type:    notify.ActivityType,
		Content: "%s '%s'",
		Args:    []any{action, displayName},
		Writer:  s.Writer,
	})
}

// generateKSailConfig generates the ksail.yaml configuration file.
func (s *Scaffolder) generateKSailConfig(output string, force bool) error {
	// Apply distribution-specific defaults to ensure consistency with generated files
	config := s.applyKSailConfigDefaults()

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, "ksail.yaml"),
		Force:  force,
	}

	return generateWithFileHandling(
		s,
		GenerationParams[v1alpha1.Cluster]{
			Gen:         s.KSailYAMLGenerator,
			Model:       config,
			Opts:        opts,
			DisplayName: "ksail.yaml",
			Force:       force,
			WrapErr: func(err error) error {
				return fmt.Errorf("%w: %w", ErrKSailConfigGeneration, err)
			},
		},
	)
}

// generateDistributionConfig generates the distribution-specific configuration file.
func (s *Scaffolder) generateDistributionConfig(output string, force bool) error {
	switch s.KSailConfig.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return s.generateKindConfig(output, force)
	case v1alpha1.DistributionK3d:
		return s.generateK3dConfig(output, force)
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

	return generateWithFileHandling(
		s,
		GenerationParams[*v1alpha4.Cluster]{
			Gen:         s.KindGenerator,
			Model:       kindConfig,
			Opts:        opts,
			DisplayName: "kind.yaml",
			Force:       force,
			WrapErr: func(err error) error {
				return fmt.Errorf("%w: %w", ErrKindConfigGeneration, err)
			},
		},
	)
}

// generateK3dConfig generates the k3d.yaml configuration file.
func (s *Scaffolder) generateK3dConfig(output string, force bool) error {
	k3dConfig := s.createK3dConfig()

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, "k3d.yaml"),
		Force:  force,
	}

	return generateWithFileHandling(
		s,
		GenerationParams[*k3dv1alpha5.SimpleConfig]{
			Gen:         s.K3dGenerator,
			Model:       &k3dConfig,
			Opts:        opts,
			DisplayName: "k3d.yaml",
			Force:       force,
			WrapErr: func(err error) error {
				return fmt.Errorf("%w: %w", ErrK3dConfigGeneration, err)
			},
		},
	)
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

// generateKustomizationConfig generates the kustomization.yaml file.
func (s *Scaffolder) generateKustomizationConfig(output string, force bool) error {
	kustomization := ktypes.Kustomization{}

	opts := yamlgenerator.Options{
		Output: filepath.Join(output, s.KSailConfig.Spec.SourceDirectory, "kustomization.yaml"),
		Force:  force,
	}

	return generateWithFileHandling(
		s,
		GenerationParams[*ktypes.Kustomization]{
			Gen:         s.KustomizationGenerator,
			Model:       &kustomization,
			Opts:        opts,
			DisplayName: "k8s/kustomization.yaml",
			Force:       force,
			WrapErr: func(err error) error {
				return fmt.Errorf("%w: %w", ErrKustomizationGeneration, err)
			},
		},
	)
}
