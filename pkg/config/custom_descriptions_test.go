package config_test

import (
	"bytes"
	"strings"
	"testing"

	v1alpha1 "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
)

// TestNewCobraCommandWithDescriptions verifies that custom flag descriptions
// can be provided when constructing Cobra commands using AddFlagsFromFields.
func TestNewCobraCommandWithDescriptions(t *testing.T) {
	t.Parallel()

	// Create command with custom descriptions using AddFlagsFromFields
	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test command with custom descriptions",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error { return nil },
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{
				&c.Spec.Distribution, "Choose your preferred Kubernetes distribution",
				&c.Spec.SourceDirectory, "Path to workload manifests",
			}
		})...,
	)

	// Capture help output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	helpOutput := out.String()

	// Verify custom descriptions are used
	if !strings.Contains(helpOutput, "Choose your preferred Kubernetes distribution") {
		t.Error("custom distribution description not found in help output")
	}

	if !strings.Contains(helpOutput, "Path to workload manifests") {
		t.Error("custom source-directory description not found in help output")
	}

	// Verify flags exist with correct names
	if !strings.Contains(helpOutput, "--distribution") {
		t.Error("distribution flag not found in help output")
	}

	if !strings.Contains(helpOutput, "--source-directory") {
		t.Error("source-directory flag not found in help output")
	}
}

// TestNewCobraCommandWithoutDescriptions verifies that the default descriptions
// are used when no custom descriptions are provided.
func TestNewCobraCommandWithoutDescriptions(t *testing.T) {
	t.Parallel()

	// Create command without custom descriptions (using AddFlagsFromFields)
	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test command with default descriptions",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error { return nil },
		config.AddFlagsFromFields(func(c *v1alpha1.Cluster) []any {
			return []any{&c.Spec.Distribution, &c.Spec.SourceDirectory}
		})...,
	)

	// Capture help output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	helpOutput := out.String()

	// Verify default descriptions are used
	if !strings.Contains(helpOutput, "Configure Distribution") {
		t.Error("default distribution description not found in help output")
	}

	if !strings.Contains(helpOutput, "Configure SourceDirectory") {
		t.Error("default source-directory description not found in help output")
	}
}

// TestNewCobraCommandMixedDescriptions verifies that mixed field selectors work correctly
// (some fields have custom descriptions, others use defaults).
func TestNewCobraCommandMixedDescriptions(t *testing.T) {
	t.Parallel()

	// Create command with mixed field selectors
	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test command with mixed descriptions",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error { return nil },
		config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			"Select Kubernetes distribution (Kind, K3d, EKS, Tind)"),
		config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory }),
	)

	// Capture help output
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--help"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	helpOutput := out.String()

	// Verify custom description is used for distribution
	if !strings.Contains(helpOutput, "Select Kubernetes distribution (Kind, K3d, EKS, Tind)") {
		t.Error("custom distribution description not found in help output")
	}

	// Verify default description is used for source-directory
	if !strings.Contains(helpOutput, "Configure SourceDirectory") {
		t.Error("default source-directory description not found in help output")
	}
}
