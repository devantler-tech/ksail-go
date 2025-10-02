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
// It prints stub messages and creates minimal empty files just so tests pass.
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

// Scaffold prints stub messages and creates minimal empty files for testing.
func (s *Scaffolder) Scaffold(output string, force bool) error {
	_, _ = fmt.Fprintf(s.Writer, "STUB: Scaffolding project to '%s' (force=%v)\n", output, force)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Distribution: %s\n", s.KSailConfig.Spec.Distribution)

	// Create minimal directory structure
	err := os.MkdirAll(output, dirPerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	err = os.MkdirAll(filepath.Join(output, "k8s"), dirPerm)
	if err != nil {
		return fmt.Errorf("failed to create k8s directory: %w", err)
	}

	// Create empty stub files just so they exist
	files := []string{
		filepath.Join(output, "ksail.yaml"),
		filepath.Join(output, string(s.KSailConfig.Spec.Distribution)+".yaml"),
		filepath.Join(output, "k8s", "kustomization.yaml"),
	}

	for _, file := range files {
		err = os.WriteFile(file, []byte("# stub\n"), filePerm)
		if err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
	}

	_, _ = fmt.Fprintf(s.Writer, "STUB: Created ksail.yaml\n")
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created %s.yaml\n", s.KSailConfig.Spec.Distribution)
	_, _ = fmt.Fprintf(s.Writer, "STUB: Created k8s/kustomization.yaml\n")

	return nil
}
