// Package scaffolder provides utilities for scaffolding KSail project files and configuration.
package scaffolder

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/k3d-io/k3d/v5/pkg/config/types"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	ktypes "sigs.k8s.io/kustomize/api/types"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/ui/notify"
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
		distributionName = "k3d-default" // Default K3d cluster name (handled by config manager)

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
	MirrorRegistries       []string // Format: "name=upstream" (e.g., "docker-io=https://registry-1.docker.io")
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
		Type:    notify.GenerateType,
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

	// Disable default CNI if Cilium is requested
	if s.KSailConfig.Spec.CNI == v1alpha1.CNICilium {
		kindConfig.Networking.DisableDefaultCNI = true
	}

	// Add containerd config patches for mirror registries
	if len(s.MirrorRegistries) > 0 {
		kindConfig.ContainerdConfigPatches = s.generateContainerdPatches()
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
		// Additional configuration will be handled by the provisioner with sensible defaults
		// Users can override any settings in this generated config file
	}

	// Disable default CNI (Flannel) if Cilium is requested
	if s.KSailConfig.Spec.CNI == v1alpha1.CNICilium {
		config.Options.K3sOptions.ExtraArgs = []k3dv1alpha5.K3sArgWithNodeFilters{
			{
				Arg:         "--flannel-backend=none",
				NodeFilters: []string{"server:*"},
			},
			{
				Arg:         "--disable-network-policy",
				NodeFilters: []string{"server:*"},
			},
		}
	}

	// Add registry configuration for mirror registries
	if len(s.MirrorRegistries) > 0 {
		config.Registries = s.generateK3dRegistryConfig()
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
			DisplayName: filepath.Join(s.KSailConfig.Spec.SourceDirectory, "kustomization.yaml"),
			Force:       force,
			WrapErr: func(err error) error {
				return fmt.Errorf("%w: %w", ErrKustomizationGeneration, err)
			},
		},
	)
}

// generateContainerdPatches generates containerd config patches for Kind mirror registries.
// Input format: "name=upstream" (e.g., "docker-io=https://registry-1.docker.io")
// Container names are generated as "kind-{name}" for Kind network DNS resolution.
func (s *Scaffolder) generateContainerdPatches() []string {
	patches := make([]string, 0, len(s.MirrorRegistries))

	for _, mirrorSpec := range s.MirrorRegistries {
		parts := splitMirrorSpec(mirrorSpec)
		if parts == nil {
			continue
		}

		name := parts[0]
		upstream := parts[1]

		// Extract port from upstream URL (default: 5000)
		port := extractPortFromURL(upstream)

		// Generate distribution-prefixed container name: kind-{name}
		containerName := "kind-" + name

		// Infer registry host from name (e.g., docker-io -> docker.io)
		registryHost := inferRegistryHost(name)

		// Use container name as endpoint for Kind network DNS resolution
		kindEndpoint := "http://" + containerName + ":" + port

		patch := fmt.Sprintf(`[plugins."io.containerd.grpc.v1.cri".registry.mirrors."%s"]
  endpoint = ["%s"]`, registryHost, kindEndpoint)

		patches = append(patches, patch)
	}

	return patches
}

// inferRegistryHost converts a registry name back to its host format.
// Examples: "docker-io" -> "docker.io", "ghcr-io" -> "ghcr.io".
func inferRegistryHost(name string) string {
	// Replace hyphens with dots to get the registry host
	// Handle common cases: docker-io -> docker.io, ghcr-io -> ghcr.io
	return strings.ReplaceAll(name, "-", ".")
}

// extractPortFromURL extracts the port from a URL string.
// Returns "5000" as default if no port is found.
func extractPortFromURL(urlStr string) string {
	// Remove protocol if present
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	// Find port after colon
	if idx := strings.LastIndex(urlStr, ":"); idx >= 0 {
		port := urlStr[idx+1:]
		// Remove any path after the port
		if slashIdx := strings.Index(port, "/"); slashIdx >= 0 {
			port = port[:slashIdx]
		}

		return port
	}

	// Default port for registry
	return "5000"
}

// generateK3dRegistryConfig generates K3d registry configuration for mirror registries.
// Input format: "name=upstream" (e.g., "docker-io=https://registry-1.docker.io")
// K3d requires one registry per proxy, so we generate multiple create configs.
func (s *Scaffolder) generateK3dRegistryConfig() k3dv1alpha5.SimpleConfigRegistries {
	registryConfig := k3dv1alpha5.SimpleConfigRegistries{
		Use: []string{},
	}

	// K3d requires one registry per upstream proxy
	// We'll create multiple registries with distribution-prefixed names: k3d-{name}
	if len(s.MirrorRegistries) > 0 {
		// For now, we'll use the first mirror as the primary registry
		// Multiple mirrors require multiple registries, which K3d supports via separate create configs
		mirrorSpec := s.MirrorRegistries[0]

		parts := splitMirrorSpec(mirrorSpec)
		if parts != nil {
			name := parts[0]
			// upstream := parts[1] // TODO: Use upstream for proxy configuration

			// Generate distribution-prefixed registry name
			registryName := "k3d-" + name

			registryConfig.Create = &k3dv1alpha5.SimpleConfigRegistryCreateConfig{
				Name: registryName,
			}

			// Generate mirrors configuration
			configLines := make([]string, 0, len(s.MirrorRegistries)*4+1)
			configLines = append(configLines, "mirrors:")

			// Infer registry host from name
			registryHost := inferRegistryHost(name)

			configLines = append(configLines, fmt.Sprintf(`  "%s":`, registryHost))
			configLines = append(configLines, "    endpoint:")
			configLines = append(configLines, fmt.Sprintf("      - http://%s:5000", registryName))

			// Add proxy configuration
			configLines = append(configLines, "configs:")
			configLines = append(configLines, fmt.Sprintf(`  "%s:5000":`, registryName))
			configLines = append(configLines, "    auth: {}")
			configLines = append(configLines, "    tls:")
			configLines = append(configLines, "      insecure_skip_verify: false")

			registryConfig.Config = joinLines(configLines) + "\n"
		}
	}

	return registryConfig
}

// splitMirrorSpec splits a mirror specification into registry and endpoint parts.
// Returns nil if the spec is invalid.
func splitMirrorSpec(spec string) []string {
	parts := splitOnEquals(spec)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}

	return parts
}

// splitOnEquals splits a string on the first '=' character.
func splitOnEquals(str string) []string {
	idx := findFirstEquals(str)
	if idx == -1 {
		return []string{str}
	}

	return []string{str[:idx], str[idx+1:]}
}

// findFirstEquals finds the index of the first '=' character.
func findFirstEquals(s string) int {
	for i, c := range s {
		if c == '=' {
			return i
		}
	}

	return -1
}

// joinLines joins strings with newlines.
func joinLines(lines []string) string {
	result := ""

	for idx, line := range lines {
		if idx > 0 {
			result += "\n"
		}

		result += line
	}

	return result
}
