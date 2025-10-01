package stubs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

const (
	stubDirPerm  = 0o750
	stubFilePerm = 0o600
)

// Scaffolder is a stub implementation of the Scaffolder for integration testing.
type Scaffolder struct {
	KSailConfig            v1alpha1.Cluster
	KSailYAMLGenerator     generator.Generator[v1alpha1.Cluster, yamlgenerator.Options]
	KindGenerator          generator.Generator[*v1alpha4.Cluster, yamlgenerator.Options]
	K3dGenerator           generator.Generator[*v1alpha5.SimpleConfig, yamlgenerator.Options]
	KustomizationGenerator generator.Generator[*ktypes.Kustomization, yamlgenerator.Options]
	Writer                 io.Writer
}

// NewScaffolder creates a new stub scaffolder.
func NewScaffolder(cfg v1alpha1.Cluster, writer io.Writer) *Scaffolder {
	return &Scaffolder{
		KSailConfig: cfg,
		Writer:      writer,
	}
}

// Scaffold simulates generating project files and configurations.
//
//nolint:funlen // Scaffold function requires multiple file operations
func (s *Scaffolder) Scaffold(output string, force bool) error {
	_, _ = fmt.Fprintf(s.Writer, "STUB: Scaffolding project to '%s' (force=%v)\n", output, force)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Distribution: %s\n", s.KSailConfig.Spec.Distribution)

	// Create minimal stub files to allow commands to find configuration
	//nolint:noinlineerr // Inline error handling is appropriate for stub code
	if err := os.MkdirAll(output, stubDirPerm); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create stub ksail.yaml with minimal valid configuration
	ksailYamlPath := filepath.Join(output, "ksail.yaml")

	// Determine the correct config file and context based on distribution
	var configFile, context string

	switch s.KSailConfig.Spec.Distribution {
	case v1alpha1.DistributionKind:
		configFile = "kind.yaml"
		context = "kind-kind"
	case v1alpha1.DistributionK3d:
		configFile = "k3d.yaml"
		context = "k3d-k3s-default"
	default:
		configFile = string(s.KSailConfig.Spec.Distribution) + ".yaml"
		context = ""
	}

	// Include context in the YAML for distributions that need it
	var stubContent string
	if context != "" {
		stubContent = fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
  connection:
    context: %s
`, s.KSailConfig.Spec.Distribution, configFile, context)
	} else {
		stubContent = fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
`, s.KSailConfig.Spec.Distribution, configFile)
	}

	//nolint:noinlineerr // Inline error handling is appropriate for stub code
	if err := os.WriteFile(ksailYamlPath, []byte(stubContent), stubFilePerm); err != nil {
		return fmt.Errorf("failed to write stub ksail.yaml: %w", err)
	}

	// Create stub distribution config file
	distConfigPath := filepath.Join(output, configFile)

	distConfig := "# Stub distribution configuration\n"
	if err := os.WriteFile(distConfigPath, []byte(distConfig), stubFilePerm); err != nil {
		return fmt.Errorf("failed to write stub distribution config: %w", err)
	}

	// Create stub k8s directory with kustomization.yaml
	k8sDir := filepath.Join(output, "k8s")
	if err := os.MkdirAll(k8sDir, stubDirPerm); err != nil {
		return fmt.Errorf("failed to create k8s directory: %w", err)
	}

	kustomizationPath := filepath.Join(k8sDir, "kustomization.yaml")

	kustomization := "# Stub kustomization\n"
	if err := os.WriteFile(kustomizationPath, []byte(kustomization), stubFilePerm); err != nil {
		return fmt.Errorf("failed to write stub kustomization: %w", err)
	}

	_, _ = fmt.Fprintf(s.Writer, "STUB: Created ksail.yaml\n")
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created %s\n", configFile)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created k8s/kustomization.yaml\n")

	return nil
}
