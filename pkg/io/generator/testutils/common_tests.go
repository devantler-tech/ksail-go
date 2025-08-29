// Package testutils provides generator-specific test utilities.
package testutils

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
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

// TestExistingFileNoForce runs a common test pattern for generators with existing files and no force flag.
func TestExistingFileNoForce[T any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	cluster T,
	filename string,
	assertContent func(*testing.T, string, string),
	clusterName string,
) {
	t.Helper()
	TestExistingFile(t, gen, cluster, filename, assertContent, clusterName, false)
}

// TestExistingFileWithForce runs a common test pattern for generators with existing files and force flag.
func TestExistingFileWithForce[T any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	cluster T,
	filename string,
	assertContent func(*testing.T, string, string),
	clusterName string,
) {
	t.Helper()
	TestExistingFile(t, gen, cluster, filename, assertContent, clusterName, true)
}