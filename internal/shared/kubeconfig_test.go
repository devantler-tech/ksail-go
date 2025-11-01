package shared_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/shared"
	cmdtestutils "github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/stretchr/testify/require"
)

func TestGetDefaultKubeconfigPath(t *testing.T) {
	t.Parallel()

	path := shared.GetDefaultKubeconfigPath()

	homeDir, _ := os.UserHomeDir()
	expected := filepath.Join(homeDir, ".kube", "config")

	require.Equal(t, expected, path, "expected default kubeconfig path")
}

func TestGetKubeconfigPathFromConfig(t *testing.T) {
	t.Parallel()

	homeDir, _ := os.UserHomeDir()

	tests := []struct {
		name           string
		kubeconfigPath string
		expectedPath   string
	}{
		{
			name:           "empty path returns default",
			kubeconfigPath: "",
			expectedPath:   filepath.Join(homeDir, ".kube", "config"),
		},
		{
			name:           "absolute path unchanged",
			kubeconfigPath: "/tmp/kubeconfig",
			expectedPath:   "/tmp/kubeconfig",
		},
		{
			name:           "tilde path expanded",
			kubeconfigPath: "~/.kube/config",
			expectedPath:   filepath.Join(homeDir, ".kube", "config"),
		},
		{
			name:           "tilde path with nested dir expanded",
			kubeconfigPath: "~/custom/kubeconfig",
			expectedPath:   filepath.Join(homeDir, "custom", "kubeconfig"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			cfg := &v1alpha1.Cluster{
				Spec: v1alpha1.Spec{
					Connection: v1alpha1.Connection{
						Kubeconfig: testCase.kubeconfigPath,
					},
				},
			}

			path, err := shared.GetKubeconfigPathFromConfig(cfg)
			require.NoError(t, err)
			require.Equal(t, testCase.expectedPath, path)
		})
	}
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
