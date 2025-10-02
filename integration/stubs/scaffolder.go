package stubs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

const (
	// File permissions for directories and files.
	dirPerm  = 0o750
	filePerm = 0o600
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
	err := os.MkdirAll(output, dirPerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine config file and context for the distribution
	configFile, context := s.getDistributionConfig()

	// Generate configuration content
	ksailContent := s.generateKSailContent(configFile, context)

	// Write all configuration files
	return s.writeConfigFiles(output, configFile, ksailContent)
}

// getDistributionConfig returns the config file name and context for the distribution.
func (s *Scaffolder) getDistributionConfig() (string, string) {
	switch s.KSailConfig.Spec.Distribution {
	case v1alpha1.DistributionKind:
		return "kind.yaml", "kind-kind"
	case v1alpha1.DistributionK3d:
		return "k3d.yaml", "k3d-k3s-default"
	default:
		return string(s.KSailConfig.Spec.Distribution) + ".yaml", ""
	}
}

// generateKSailContent generates the ksail.yaml content.
func (s *Scaffolder) generateKSailContent(configFile, context string) string {
	if context != "" {
		return fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
  connection:
    context: %s
`, s.KSailConfig.Spec.Distribution, configFile, context)
	}

	return fmt.Sprintf(`apiVersion: ksail.dev/v1alpha1
kind: Cluster
spec:
  distribution: %s
  distributionConfig: %s
`, s.KSailConfig.Spec.Distribution, configFile)
}

// writeConfigFiles writes all configuration files to disk.
func (s *Scaffolder) writeConfigFiles(output, configFile, ksailContent string) error {
	// Write ksail.yaml
	err := os.WriteFile(filepath.Join(output, "ksail.yaml"), []byte(ksailContent), filePerm)
	if err != nil {
		return fmt.Errorf("failed to write ksail.yaml: %w", err)
	}

	// Write distribution config
	distContent := "# Stub distribution configuration\n"

	err = os.WriteFile(filepath.Join(output, configFile), []byte(distContent), filePerm)
	if err != nil {
		return fmt.Errorf("failed to write distribution config: %w", err)
	}

	// Create k8s directory and kustomization
	err = s.writeKustomization(output)
	if err != nil {
		return err
	}

	_, _ = fmt.Fprintf(s.Writer, "STUB: Created ksail.yaml\n")
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created %s\n", configFile)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created k8s/kustomization.yaml\n")

	return nil
}

// writeKustomization creates the k8s directory and writes kustomization.yaml.
func (s *Scaffolder) writeKustomization(output string) error {
	k8sDir := filepath.Join(output, "k8s")

	err := os.MkdirAll(k8sDir, dirPerm)
	if err != nil {
		return fmt.Errorf("failed to create k8s directory: %w", err)
	}

	kustomizationContent := "# Stub kustomization\n"
	kustomizationPath := filepath.Join(k8sDir, "kustomization.yaml")

	err = os.WriteFile(kustomizationPath, []byte(kustomizationContent), filePerm)
	if err != nil {
		return fmt.Errorf("failed to write kustomization.yaml: %w", err)
	}

	return nil
}
