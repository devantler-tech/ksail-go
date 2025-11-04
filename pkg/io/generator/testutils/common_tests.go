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
	Force       bool
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
				Force:  testCase.Force,
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

// GetStandardGenerateTestCasesWithForce returns standard test cases including force overwrite.
func GetStandardGenerateTestCasesWithForce(expectedFileName string) []GenerateTestCase {
	standardCases := GetStandardGenerateTestCases(expectedFileName)

	forceCase := GenerateTestCase{
		Name:        "with force overwrite",
		ClusterName: "force-cluster",
		Force:       true,
		SetupOutput: func(t *testing.T) (string, bool, string) {
			t.Helper()
			tempDir := t.TempDir()
			outputPath := filepath.Join(tempDir, expectedFileName)

			// Create existing file first
			err := ioutils.WriteFileSafe("existing content", tempDir, outputPath, true)
			require.NoError(t, err)

			return outputPath, true, tempDir
		},
		ExpectError: false,
	}

	return append(standardCases, forceCase)
}

// RunStandardGeneratorTests runs the standard generator test suite.
func RunStandardGeneratorTests[T any](
	t *testing.T,
	gen generator.Generator[T, yamlgenerator.Options],
	createCluster func(name string) T,
	expectedFileName string,
	assertContent func(*testing.T, string, string),
) {
	t.Helper()

	testCases := GetStandardGenerateTestCasesWithForce(expectedFileName)

	TestGenerateCommon(
		t,
		testCases,
		createCluster,
		gen,
		assertContent,
		expectedFileName,
	)
}
