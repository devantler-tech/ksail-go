// Package testutils provides generator-specific test utilities.
package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExistingFile runs a common test pattern for generators with existing files.
func TestExistingFile[T any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	cluster T,
	filename string,
	assertContent func(*testing.T, string, string),
	clusterName string,
	force bool,
) {
	t.Helper()

	// Arrange
	tempDir, outputPath, existingContent := testutils.SetupExistingFile(t, filename)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  force,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertContent(t, result, clusterName)

	if force {
		// Verify file was overwritten
		testutils.AssertFileEquals(t, tempDir, outputPath, result)

		// Additional check: ensure old content was replaced
		fileContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
		require.NoError(t, err, "File should exist")
		assert.NotEqual(t, existingContent, string(fileContent), "Old content should be replaced")
	} else {
		// Verify file was NOT overwritten
		testutils.AssertFileEquals(t, tempDir, outputPath, existingContent)
	}
}

// TestFileWriteError runs a common test pattern for generators with file write errors.
func TestFileWriteError[T any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	cluster T,
	filename string,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange - Use an invalid file path that will cause a write error
	invalidPath := "/dev/null/invalid/path/" + filename
	opts := yamlgenerator.Options{
		Output: invalidPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), expectedErrorContains, "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}
