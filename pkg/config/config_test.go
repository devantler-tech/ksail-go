package config_test

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/config"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestManager_LoadCluster_Defaults(t *testing.T) {
	t.Parallel()

	// Clear any existing environment variables that might affect the test
	envVarsToClean := []string{
		"KSAIL_SPEC_DISTRIBUTION",
		"KSAIL_METADATA_NAME",
		"KSAIL_SPEC_CONNECTION_KUBECONFIG",
	}
	for _, envVar := range envVarsToClean {
		if originalValue := os.Getenv(envVar); originalValue != "" {
			_ = os.Unsetenv(envVar)
			defer func(envVar, originalValue string) {
				_ = os.Setenv(envVar, originalValue)
			}(envVar, originalValue)
		}
	}

	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()

	defer func() { _ = os.Chdir(oldDir) }()

	_ = os.Chdir(tempDir)

	// Load cluster with field selectors to provide defaults
	fieldSelectors := []config.FieldSelector[v1alpha1.Cluster]{
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"ksail-default",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.DistributionConfig },
			"kind.yaml",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			"~/.kube/config",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"kind-ksail-default",
		),
	}
	manager := config.NewManager(fieldSelectors...)
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test defaults
	assert.Equal(t, "ksail-default", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionKind, cluster.Spec.Distribution)
	assert.Equal(t, "kind.yaml", cluster.Spec.DistributionConfig)
	assert.Equal(t, "k8s", cluster.Spec.SourceDirectory)
	assert.Equal(t, "~/.kube/config", cluster.Spec.Connection.Kubeconfig)
	assert.Equal(t, "kind-ksail-default", cluster.Spec.Connection.Context)
}

func TestManager_LoadCluster_EnvironmentVariables(t *testing.T) {
	t.Parallel()

	// Set environment variables - using the correct hierarchical structure
	_ = os.Setenv("KSAIL_METADATA_NAME", "test-cluster")
	_ = os.Setenv("KSAIL_SPEC_DISTRIBUTION", "K3d")
	_ = os.Setenv("KSAIL_SPEC_CONNECTION_KUBECONFIG", "/custom/path/kubeconfig")

	defer func() {
		_ = os.Unsetenv("KSAIL_METADATA_NAME")
		_ = os.Unsetenv("KSAIL_SPEC_DISTRIBUTION")
		_ = os.Unsetenv("KSAIL_SPEC_CONNECTION_KUBECONFIG")
	}()

	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()

	defer func() { _ = os.Chdir(oldDir) }()

	_ = os.Chdir(tempDir)

	fieldSelectors := []config.FieldSelector[v1alpha1.Cluster]{
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"ksail-default",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			"~/.kube/config",
		),
	}
	manager := config.NewManager(fieldSelectors...)
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Environment variables should override defaults
	assert.Equal(t, "test-cluster", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
	assert.Equal(t, "/custom/path/kubeconfig", cluster.Spec.Connection.Kubeconfig)
}

func TestNewManager(t *testing.T) {
	t.Parallel()

	manager := config.NewManager()
	require.NotNil(t, manager)
	require.NotNil(t, manager.GetViper())
}

func TestManager_LoadCluster_ConfigFile(t *testing.T) {
	t.Parallel()

	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()

	defer func() { _ = os.Chdir(oldDir) }()

	_ = os.Chdir(tempDir)

	// Create a ksail.yaml config file
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: config-test-cluster
spec:
  distribution: K3d
  sourceDirectory: config-k8s
  connection:
    kubeconfig: /config/path/kubeconfig
    context: config-context
    timeout: 60s
`
	err := os.WriteFile("ksail.yaml", []byte(configContent), 0o600)
	require.NoError(t, err)

	fieldSelectors := []config.FieldSelector[v1alpha1.Cluster]{
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"ksail-default",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			"~/.kube/config",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"kind-ksail-default",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Timeout },
			metav1.Duration{Duration: 5 * time.Minute},
		),
	}
	manager := config.NewManager(fieldSelectors...)
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test config file values are loaded
	assert.Equal(t, "config-test-cluster", cluster.Metadata.Name)
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)
	assert.Equal(t, "config-k8s", cluster.Spec.SourceDirectory)
	assert.Equal(t, "/config/path/kubeconfig", cluster.Spec.Connection.Kubeconfig)
	assert.Equal(t, "config-context", cluster.Spec.Connection.Context)
	assert.Equal(t, metav1.Duration{Duration: 60 * time.Second}, cluster.Spec.Connection.Timeout)
}

func TestManager_LoadCluster_MixedConfiguration(t *testing.T) {
	t.Parallel()

	// Setup a temporary directory for testing
	tempDir := t.TempDir()
	oldDir, _ := os.Getwd()

	defer func() { _ = os.Chdir(oldDir) }()

	_ = os.Chdir(tempDir)

	// Create a ksail.yaml config file with some values
	configContent := `apiVersion: ksail.dev/v1alpha1
kind: Cluster
metadata:
  name: config-cluster
spec:
  distribution: K3d
  sourceDirectory: config-k8s
  connection:
    kubeconfig: /config/path/kubeconfig
    context: config-context
`
	err := os.WriteFile("ksail.yaml", []byte(configContent), 0o600)
	require.NoError(t, err)

	// Set environment variables (should override config file)
	_ = os.Setenv("KSAIL_METADATA_NAME", "env-cluster")
	_ = os.Setenv("KSAIL_SPEC_CONNECTION_KUBECONFIG", "/env/path/kubeconfig")

	defer func() {
		_ = os.Unsetenv("KSAIL_METADATA_NAME")
		_ = os.Unsetenv("KSAIL_SPEC_CONNECTION_KUBECONFIG")
	}()

	fieldSelectors := []config.FieldSelector[v1alpha1.Cluster]{
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Metadata.Name },
			"ksail-default",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
			v1alpha1.DistributionKind,
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Kubeconfig },
			"~/.kube/config",
		),
		config.AddFlagFromField(
			func(c *v1alpha1.Cluster) any { return &c.Spec.Connection.Context },
			"kind-ksail-default",
		),
	}
	manager := config.NewManager(fieldSelectors...)
	cluster, err := manager.LoadCluster()
	require.NoError(t, err)

	// Test precedence: env vars override config file
	assert.Equal(t, "env-cluster", cluster.Metadata.Name)                       // From env var
	assert.Equal(t, "/env/path/kubeconfig", cluster.Spec.Connection.Kubeconfig) // From env var
	assert.Equal(t, v1alpha1.DistributionK3d, cluster.Spec.Distribution)        // From config file
	assert.Equal(t, "config-k8s", cluster.Spec.SourceDirectory)                 // From config file
	assert.Equal(t, "config-context", cluster.Spec.Connection.Context)          // From config file
}

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
				&c.Spec.Distribution, v1alpha1.DistributionKind, "Choose your preferred Kubernetes distribution",
				&c.Spec.SourceDirectory, "k8s", "Path to workload manifests",
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
			return []any{
				&c.Spec.Distribution, v1alpha1.DistributionKind,
				&c.Spec.SourceDirectory, "k8s",
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

	// Verify default descriptions are used (empty since no descriptions provided)
	if !strings.Contains(helpOutput, "--distribution") {
		t.Error("distribution flag not found in help output")
	}

	if !strings.Contains(helpOutput, "--source-directory") {
		t.Error("source-directory flag not found in help output")
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
			v1alpha1.DistributionKind, "Select Kubernetes distribution (Kind, K3d, EKS, Tind)"),
		config.AddFlagFromField(func(c *v1alpha1.Cluster) any { return &c.Spec.SourceDirectory },
			"k8s"),
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

	// Verify flags exist
	if !strings.Contains(helpOutput, "--distribution") {
		t.Error("distribution flag not found in help output")
	}

	if !strings.Contains(helpOutput, "--source-directory") {
		t.Error("source-directory flag not found in help output")
	}
}
