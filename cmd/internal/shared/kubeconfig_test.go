package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/internal/shared"
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
