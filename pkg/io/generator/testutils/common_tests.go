package testutils

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/io/generator"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// GenerateTestCase represents a test case for the basic TestGenerate pattern.
type GenerateTestCase struct {
	Name        string
	ClusterName string
	SetupOutput func(t *testing.T) (output string, verifyFile bool, tempDir string)
	ExpectError bool
}

// TestGenerateCommon runs the common TestGenerate pattern with shared setup and assertions.
func TestGenerateCommon[T any]( //nolint:tparallel
	t *testing.T,
	tests []GenerateTestCase,
	createCluster func(name string) T,
	gen generator.Generator[T, yamlgenerator.Options],
	assertContent func(*testing.T, string, string),
	expectedFileName string,
) {
	t.Helper()

	for _, testCase := range tests {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			cluster := createCluster(testCase.ClusterName)
			output, verifyFile, tempDir := testCase.SetupOutput(t)
			opts := yamlgenerator.Options{
				Output: output,
			}

			result, err := gen.Generate(cluster, opts)

			if testCase.ExpectError {
				require.Error(t, err)
				assert.Empty(t, result)
			} else {
				require.NoError(t, err)
				assertContent(t, result, testCase.ClusterName)

				if verifyFile {
					testutils.AssertFileEquals(t, tempDir, filepath.Join(tempDir, expectedFileName), result)
				}
			}
		})
	}
}

// GetStandardGenerateTestCases returns standard test cases for generator testing.
func GetStandardGenerateTestCases(expectedFileName string) []GenerateTestCase {
	return []GenerateTestCase{
		{
			Name:        "without file",
			ClusterName: "test-cluster",
			SetupOutput: func(_ *testing.T) (string, bool, string) {
				return "", false, ""
			},
			ExpectError: false,
		},
		{
			Name:        "with file",
			ClusterName: "file-cluster",
			SetupOutput: func(t *testing.T) (string, bool, string) {
				t.Helper()
				tempDir := t.TempDir()
				outputPath := filepath.Join(tempDir, expectedFileName)

				return outputPath, true, tempDir
			},
			ExpectError: false,
		},
	}
}

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

	tempDir, outputPath, existingContent := testutils.SetupExistingFile(t, filename)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  force,
	}

	result, err := gen.Generate(cluster, opts)

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

	result, err := gen.Generate(cluster, opts)

	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), expectedErrorContains, "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}
