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
// can be provided when constructing Cobra commands.
func TestNewCobraCommandWithDescriptions(t *testing.T) {
	t.Parallel()

	// Define custom descriptions
	customDescriptions := map[string]string{
		"distribution":      "Choose your preferred Kubernetes distribution",
		"source-directory": "Path to workload manifests",
	}

	// Create command with custom descriptions
	cmd := config.NewCobraCommandWithDescriptions(
		"test",
		"Test command",
		"Test command with custom descriptions",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error { return nil },
		customDescriptions,
		config.Fields(func(c *v1alpha1.Cluster) []any {
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

	// Create command without custom descriptions (using original constructor)
	cmd := config.NewCobraCommand(
		"test",
		"Test command",
		"Test command with default descriptions",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error { return nil },
		config.Fields(func(c *v1alpha1.Cluster) []any {
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

// TestNewCobraCommandPartialDescriptions verifies that partial custom descriptions
// work correctly (some flags have custom descriptions, others use defaults).
func TestNewCobraCommandPartialDescriptions(t *testing.T) {
	t.Parallel()

	// Define partial custom descriptions (only distribution)
	customDescriptions := map[string]string{
		"distribution": "Select Kubernetes distribution (Kind, K3d, EKS, Tind)",
	}

	// Create command with partial custom descriptions
	cmd := config.NewCobraCommandWithDescriptions(
		"test",
		"Test command",
		"Test command with partial custom descriptions",
		func(_ *cobra.Command, _ *config.Manager, _ []string) error { return nil },
		customDescriptions,
		config.Fields(func(c *v1alpha1.Cluster) []any {
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

	// Verify custom description is used for distribution
	if !strings.Contains(helpOutput, "Select Kubernetes distribution (Kind, K3d, EKS, Tind)") {
		t.Error("custom distribution description not found in help output")
	}

	// Verify default description is used for source-directory
	if !strings.Contains(helpOutput, "Configure SourceDirectory") {
		t.Error("default source-directory description not found in help output")
	}
}