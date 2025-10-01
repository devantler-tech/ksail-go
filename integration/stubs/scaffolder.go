package stubs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

// Scaffolder is a stub implementation of the Scaffolder for integration testing.
// It creates minimal valid configuration files without using complex generators.
type Scaffolder struct {
	KSailConfig v1alpha1.Cluster
	Writer      io.Writer
}

// NewScaffolder creates a new stub scaffolder.
func NewScaffolder(cfg v1alpha1.Cluster, writer io.Writer) *Scaffolder {
	return &Scaffolder{
		KSailConfig: cfg,
		Writer:      writer,
	}
}

// Scaffold creates minimal stub files for integration testing.
// Files are created with minimal content to allow commands to load configuration.
func (s *Scaffolder) Scaffold(output string, force bool) error {
	_, _ = fmt.Fprintf(s.Writer, "STUB: Scaffolding project to '%s' (force=%v)\n", output, force)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Distribution: %s\n", s.KSailConfig.Spec.Distribution)

	// Create output directory
	if err := os.MkdirAll(output, 0o750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine config file and context for the distribution
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

	// Create minimal ksail.yaml
	ksailContent := fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
`, s.KSailConfig.Spec.Distribution, configFile)
	if context != "" {
		ksailContent = fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
  connection:
    context: %s
`, s.KSailConfig.Spec.Distribution, configFile, context)
	}

	if err := os.WriteFile(filepath.Join(output, "ksail.yaml"), []byte(ksailContent), 0o600); err != nil {
		return fmt.Errorf("failed to write ksail.yaml: %w", err)
	}

	// Create minimal distribution config file
	distContent := "# Stub distribution configuration\n"
	if err := os.WriteFile(filepath.Join(output, configFile), []byte(distContent), 0o600); err != nil {
		return fmt.Errorf("failed to write distribution config: %w", err)
	}

	// Create k8s directory with minimal kustomization.yaml
	k8sDir := filepath.Join(output, "k8s")
	if err := os.MkdirAll(k8sDir, 0o750); err != nil {
		return fmt.Errorf("failed to create k8s directory: %w", err)
	}

	kustomizationContent := "# Stub kustomization\n"
	if err := os.WriteFile(filepath.Join(k8sDir, "kustomization.yaml"), []byte(kustomizationContent), 0o600); err != nil {
		return fmt.Errorf("failed to write kustomization.yaml: %w", err)
	}

	_, _ = fmt.Fprintf(s.Writer, "STUB: Created ksail.yaml\n")
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created %s\n", configFile)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created k8s/kustomization.yaml\n")

	return nil
}
