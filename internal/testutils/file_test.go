package testutils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestAssertFileEquals(t *testing.T) {
	t.Parallel()

	t.Run("matching_file_content", func(t *testing.T) {
		t.Parallel()

		// Setup
		tempDir := t.TempDir()
		testFile := filepath.Join(tempDir, "test.txt")
		expectedContent := "test content"
		err := os.WriteFile(testFile, []byte(expectedContent), 0o600)
		require.NoError(t, err)

		// Test - should not panic
		testutils.AssertFileEquals(t, tempDir, testFile, expectedContent)
	})
}

func TestSetupExistingFile(t *testing.T) {
	t.Parallel()

	t.Run("creates_file_with_content", func(t *testing.T) {
		t.Parallel()

		filename := "test-config.yaml"
		dir, path, content := testutils.SetupExistingFile(t, filename)

		// Verify the directory exists
		require.DirExists(t, dir)

		// Verify the file exists and has the expected content
		require.FileExists(t, path)
		data, err := os.ReadFile(path)
		require.NoError(t, err)
		require.Equal(t, content, string(data))
		require.Equal(t, "# existing content", content)

		// Verify the path is correct
		expectedPath := filepath.Join(dir, filename)
		require.Equal(t, expectedPath, path)
	})
}
