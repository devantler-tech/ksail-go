package testutils_test

import (
	"os"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/installer/testutils"
	"github.com/stretchr/testify/require"
)

func TestCreateMalformedKubeconfigFile(t *testing.T) {
	t.Parallel()

	t.Run("creates_malformed_kubeconfig", func(t *testing.T) {
		t.Parallel()

		path := testutils.CreateMalformedKubeconfigFile(t)

		// Verify file exists
		require.FileExists(t, path)

		// Read and verify content is malformed
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Contains(t, string(content), "this is not valid yaml")
		require.Contains(t, string(content), "[")
	})
}

func TestCreateEmptyKubeconfigFile(t *testing.T) {
	t.Parallel()

	t.Run("creates_empty_kubeconfig", func(t *testing.T) {
		t.Parallel()

		path := testutils.CreateEmptyKubeconfigFile(t)

		// Verify file exists
		require.FileExists(t, path)

		// Read and verify content is empty
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Empty(t, string(content))
	})
}

func TestCreateValidKubeconfigFile(t *testing.T) {
	t.Parallel()

	t.Run("creates_valid_kubeconfig", func(t *testing.T) {
		t.Parallel()

		path := testutils.CreateValidKubeconfigFile(t)

		// Verify file exists
		require.FileExists(t, path)

		// Read and verify content has expected structure
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		contentStr := string(content)
		require.Contains(t, contentStr, "apiVersion: v1")
		require.Contains(t, contentStr, "kind: Config")
		require.Contains(t, contentStr, "test-cluster")
		require.Contains(t, contentStr, "test-context")
		require.Contains(t, contentStr, "test-user")
	})
}
