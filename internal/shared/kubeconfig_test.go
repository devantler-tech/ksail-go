package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/shared"
	cmdtestutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultKubeconfigPath(t *testing.T) {
	t.Parallel()

	path := shared.GetDefaultKubeconfigPath()

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".kube", "config")

	require.Equal(t, expected, path, "expected default kubeconfig path")
}

func TestGetKubeconfigPathSilently(t *testing.T) {
	t.Parallel()

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
	tempDir := t.TempDir()
	cmdtestutils.WriteValidKsailConfig(t, tempDir)
	t.Chdir(tempDir)

	path := shared.GetKubeconfigPathSilently()

	require.True(t, filepath.IsAbs(path), "expected absolute path")
}

//nolint:paralleltest // Uses t.Chdir which is incompatible with parallel tests.
func TestGetKubeconfigPathSilentlyWithMissingConfig(t *testing.T) {
	tempDir := t.TempDir()
	t.Chdir(tempDir)

	path := shared.GetKubeconfigPathSilently()
	require.True(t, filepath.IsAbs(path), "expected absolute path")
}
