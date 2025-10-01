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
func (s *Scaffolder) Scaffold(output string, force bool) error {
	fmt.Fprintf(s.Writer, "STUB: Scaffolding project to '%s' (force=%v)\n", output, force)
	fmt.Fprintf(s.Writer, "STUB: Distribution: %s\n", s.KSailConfig.Spec.Distribution)
	
	// Create minimal stub files to allow commands to find configuration
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create stub ksail.yaml with minimal valid configuration
	ksailYamlPath := filepath.Join(output, "ksail.yaml")
	
	// Determine the correct config file based on distribution
	var configFile string
	switch s.KSailConfig.Spec.Distribution {
	case "Kind":
		configFile = "kind.yaml"
	case "K3d":
		configFile = "k3d.yaml"
	default:
		configFile = string(s.KSailConfig.Spec.Distribution) + ".yaml"
	}
	
	stubContent := fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
`, s.KSailConfig.Spec.Distribution, configFile)
	
	if err := os.WriteFile(ksailYamlPath, []byte(stubContent), 0o644); err != nil {
		return fmt.Errorf("failed to write stub ksail.yaml: %w", err)
	}

	// Create stub distribution config file
	distConfigPath := filepath.Join(output, configFile)
	distConfig := "# Stub distribution configuration\n"
	if err := os.WriteFile(distConfigPath, []byte(distConfig), 0o644); err != nil {
		return fmt.Errorf("failed to write stub distribution config: %w", err)
	}

	// Create stub k8s directory with kustomization.yaml
	k8sDir := filepath.Join(output, "k8s")
	if err := os.MkdirAll(k8sDir, 0o755); err != nil {
		return fmt.Errorf("failed to create k8s directory: %w", err)
	}

	kustomizationPath := filepath.Join(k8sDir, "kustomization.yaml")
	kustomization := "# Stub kustomization\n"
	if err := os.WriteFile(kustomizationPath, []byte(kustomization), 0o644); err != nil {
		return fmt.Errorf("failed to write stub kustomization: %w", err)
	}

	fmt.Fprintf(s.Writer, "STUB: Created ksail.yaml\n")
	fmt.Fprintf(s.Writer, "STUB: Created %s\n", configFile)
	fmt.Fprintf(s.Writer, "STUB: Created k8s/kustomization.yaml\n")

	return nil
}
