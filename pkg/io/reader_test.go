package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests are intentionally minimal and explicit to keep coverage high and behavior clear.
func TestReadFileSafe(t *testing.T) {
	t.Parallel()

	t.Run("normal read", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		filePath := filepath.Join(base, "file.txt")
		want := "hello safe"
		err := os.WriteFile(filePath, []byte(want), 0o600)
		require.NoError(t, err, "WriteFile setup")

		got, err := ioutils.ReadFileSafe(base, filePath)

		require.NoError(t, err, "ReadFileSafe")
		assert.Equal(t, want, string(got), "content")
	})

	t.Run("outside base", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		outside := filepath.Join(os.TempDir(), "outside-test-file.txt")
		err := os.WriteFile(outside, []byte("nope"), 0o600)
		require.NoError(t, err, "WriteFile setup")

		_, err = ioutils.ReadFileSafe(base, outside)

		testutils.AssertErrWrappedContains(t, err, ioutils.ErrPathOutsideBase, "", "ReadFileSafe")
	})

	t.Run("traversal attempt", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		parent := filepath.Join(base, "..", "traversal.txt")
		absParent, _ := filepath.Abs(parent)
		err := os.WriteFile(absParent, []byte("traversal"), 0o600)
		require.NoError(t, err, "WriteFile setup parent")

		attempt := filepath.Join(base, "..", "traversal.txt")

		_, err = ioutils.ReadFileSafe(base, attempt)

		testutils.AssertErrWrappedContains(t, err, ioutils.ErrPathOutsideBase, "", "ReadFileSafe")
	})

	t.Run("missing file inside base", func(t *testing.T) {
		t.Parallel()

		base := t.TempDir()
		missing := filepath.Join(base, "missing.txt")

		_, err := ioutils.ReadFileSafe(base, missing)

		testutils.AssertErrContains(t, err, "failed to read file", "ReadFileSafe")
	})
}

func TestFindFile(t *testing.T) {
	t.Parallel()

	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		absolutePath := filepath.Join(tempDir, "config.yaml")
		err := os.WriteFile(absolutePath, []byte("test"), 0o600)
		require.NoError(t, err)

		resolved, err := ioutils.FindFile(absolutePath)

		require.NoError(t, err)
		assert.Equal(t, absolutePath, resolved)
	})

	t.Run("relative path found in current directory", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory and change to it
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		// Create config file in current directory
		configFile := "test-config.yaml"
		err = os.WriteFile(configFile, []byte("test"), 0o600)
		require.NoError(t, err)

		resolved, err := ioutils.FindFile(configFile)

		require.NoError(t, err)
		expectedPath := filepath.Join(tempDir, configFile)
		assert.Equal(t, expectedPath, resolved)
	})

	t.Run("relative path found in parent directory", func(t *testing.T) {
		t.Parallel()

		// Create a temporary directory structure
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)

		// Create config file in temp directory
		configFile := "parent-config.yaml"
		configPath := filepath.Join(tempDir, configFile)
		err = os.WriteFile(configPath, []byte("test"), 0o600)
		require.NoError(t, err)

		// Create subdirectory and change to it
		subDir := filepath.Join(tempDir, "subdir")
		err = os.Mkdir(subDir, 0o750)
		require.NoError(t, err)
		err = os.Chdir(subDir)
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		resolved, err := ioutils.FindFile(configFile)

		require.NoError(t, err)
		assert.Equal(t, configPath, resolved)
	})

	t.Run("relative path not found", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		configFile := "non-existent-config.yaml"

		resolved, err := ioutils.FindFile(configFile)

		require.NoError(t, err)
		// Should return original path when not found
		assert.Equal(t, configFile, resolved)
	})

	t.Run("relative path traversal multiple levels", func(t *testing.T) {
		t.Parallel()

		// Create a deep directory structure
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)

		// Create config file at root level
		configFile := "deep-config.yaml"
		configPath := filepath.Join(tempDir, configFile)
		err = os.WriteFile(configPath, []byte("test"), 0o600)
		require.NoError(t, err)

		// Create nested subdirectories
		deepDir := filepath.Join(tempDir, "level1", "level2", "level3")
		err = os.MkdirAll(deepDir, 0o750)
		require.NoError(t, err)
		err = os.Chdir(deepDir)
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir)
		}()

		resolved, err := ioutils.FindFile(configFile)

		require.NoError(t, err)
		assert.Equal(t, configPath, resolved)
	})
}
