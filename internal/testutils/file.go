// Package testutils provides generic file-related test utilities.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// File permissions for temporary test files.
const testFileMode = 0o600

// AssertFileEquals compares the file content with the expected string.
func AssertFileEquals(t *testing.T, dir, path, expected string) {
	t.Helper()

	fileContent, err := ioutils.ReadFileSafe(dir, path)

	require.NoError(t, err, "File should exist")
	assert.Equal(t, expected, string(fileContent))
}

// SetupExistingFile creates a temporary directory and an existing config file
// with default placeholder content, returning the directory, file path, and content string.
func SetupExistingFile(t *testing.T, filename string) (string, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, filename)
	existingContent := "# existing content"
	err := os.WriteFile(outputPath, []byte(existingContent), testFileMode)
	require.NoError(t, err, "Setup: create existing file")

	return tempDir, outputPath, existingContent
}