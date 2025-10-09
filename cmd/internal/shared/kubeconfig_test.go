package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
	cmdtestutils "github.com/devantler-tech/ksail-go/cmd/internal/testutils"
	"github.com/stretchr/testify/require"
)

// setupConfigFile is a helper function that writes a config file to a temp directory
// and changes to that directory.
func setupConfigFile(t *testing.T, configContent string) {
	t.Helper()

	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "ksail.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	t.Chdir(tempDir)
}

func TestGetDefaultKubeconfigPath(t *testing.T) {
	t.Parallel()

	path := shared.GetDefaultKubeconfigPath()

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".kube", "config")

	require.Equal(t, expected, path, "expected default kubeconfig path")
}

func TestGetKubeconfigPathSilently(t *testing.T) {
	t.Parallel()

	// This test just ensures the function doesn't panic and returns a valid path
	path := shared.GetKubeconfigPathSilently()

	require.NotEmpty(t, path, "expected non-empty kubeconfig path")
	require.True(
		t,
		filepath.IsAbs(path),
		"expected absolute path, got %s",
		path,
	)
}

//nolint:paralleltest // Uses t.Chdir which is incompatible with parallel tests.
func TestGetKubeconfigPathSilentlyWithValidConfig(t *testing.T) {
	// Create temp directory with valid ksail config
	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)

	// Change to temp directory so config can be found
	t.Chdir(tempDir)

	path := shared.GetKubeconfigPathSilently()

	require.NotEmpty(t, path, "expected non-empty kubeconfig path")
	require.True(t, filepath.IsAbs(path), "expected absolute path")
}

//nolint:paralleltest // Uses t.Chdir which is incompatible with parallel tests.
func TestGetKubeconfigPathSilentlyWithNoConfig(t *testing.T) {
	// Create temp directory with no config
	tempDir := t.TempDir()

	// Change to temp directory so no config can be found
	t.Chdir(tempDir)

	path := shared.GetKubeconfigPathSilently()

	// Should fall back to default kubeconfig path
	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".kube", "config")

	require.Equal(t, expected, path, "expected default kubeconfig path when config not found")
}

//nolint:paralleltest // Uses t.Chdir which is incompatible with parallel tests.
func TestGetKubeconfigPathSilentlyWithEmptyKubeconfigInConfig(t *testing.T) {
	// Write a minimal ksail.yaml with empty kubeconfig
	configContent := `apiVersion: v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  distribution: Kind
  connection:
    kubeconfig: ""
`
	setupConfigFile(t, configContent)

	path := shared.GetKubeconfigPathSilently()

	// Should use default kubeconfig path when config has empty kubeconfig
	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".kube", "config")

	require.Equal(
		t,
		expected,
		path,
		"expected default kubeconfig path when config has empty kubeconfig",
	)
}

//nolint:paralleltest // Uses t.Chdir which is incompatible with parallel tests.
func TestGetKubeconfigPathSilentlyWithTildePath(t *testing.T) {
	// Write a ksail.yaml with tilde in kubeconfig path
	configContent := `apiVersion: v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  distribution: Kind
  connection:
    kubeconfig: "~/.kube/custom-config"
`
	setupConfigFile(t, configContent)

	path := shared.GetKubeconfigPathSilently()

	// Should expand the tilde to home directory
	// Note: The actual behavior depends on whether the config can be loaded
	// If it can be loaded, the tilde will be expanded
	// If not (e.g., missing required fields), it falls back to default
	require.NotEmpty(t, path, "expected non-empty kubeconfig path")
	require.True(t, filepath.IsAbs(path), "expected absolute path")
	require.NotContains(t, path, "~", "tilde should be expanded or default used")
}
