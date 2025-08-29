// Package testutils provides common test utilities for generator packages.
package testutils

import (
	"os"
	"path/filepath"
	"testing"

	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// File permissions for temporary test files.
const testFileMode = 0o600

// Generator is a generic interface for any YAML generator that can be tested with these utilities.
type Generator[T any] interface {
	Generate(cluster T, opts yamlgenerator.Options) (string, error)
}

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

// TestExistingFileNoForce runs a common test pattern for generators with existing files and no force flag.
func TestExistingFileNoForce[T any](
	t *testing.T,
	gen Generator[T],
	cluster T,
	filename string,
	assertContent func(*testing.T, string, string),
	clusterName string,
) {
	t.Helper()

	// Arrange
	tempDir, outputPath, existingContent := SetupExistingFile(t, filename)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertContent(t, result, clusterName)

	// Verify file was NOT overwritten
	AssertFileEquals(t, tempDir, outputPath, existingContent)
}

// TestExistingFileWithForce runs a common test pattern for generators with existing files and force flag.
func TestExistingFileWithForce[T any](
	t *testing.T,
	gen Generator[T],
	cluster T,
	filename string,
	assertContent func(*testing.T, string, string),
	clusterName string,
) {
	t.Helper()

	// Arrange
	tempDir, outputPath, existingContent := SetupExistingFile(t, filename)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertContent(t, result, clusterName)

	// Verify file was overwritten
	AssertFileEquals(t, tempDir, outputPath, result)

	// Additional check: ensure old content was replaced
	fileContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "File should exist")
	assert.NotEqual(t, existingContent, string(fileContent), "Old content should be replaced")
}