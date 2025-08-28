package yamlgenerator_test

import (
	"os"
	"path/filepath"
	"testing"

	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModel is a simple model for testing YAML generation.
type TestModel struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Enabled bool   `yaml:"enabled"`
}

func TestYAMLGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := createTestModel("1.0.0", true)
	opts := generator.Options{
		Output: "",
		Force:  false,
	} // No output file specified

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertYAMLContent(t, result, model)
}

func TestYAMLGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := createTestModel("2.0.0", false)
	tempDir, outputPath := setupTempFile(t, "test.yaml")
	opts := generator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertYAMLContent(t, result, model)
	verifyFileContent(t, tempDir, outputPath, result)
}

func TestYAMLGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen, model, tempDir, outputPath, existingContent, opts := setupExistingFileTest(t, "3.0.0", true, false)

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertYAMLContent(t, result, model)
	verifyFileContent(t, tempDir, outputPath, existingContent) // File should not be overwritten
}

func TestYAMLGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen, model, tempDir, outputPath, existingContent, opts := setupExistingFileTest(t, "4.0.0", false, true)

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertYAMLContent(t, result, model)
	verifyFileContent(t, tempDir, outputPath, result) // File should be overwritten

	// Additional check: ensure old content was replaced
	fileContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "File should exist")
	assert.NotEqual(t, existingContent, string(fileContent), "Old content should be replaced")
}

func TestYAMLGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := createTestModel("5.0.0", true)

	// Use an invalid file path that will cause a write error
	// On Unix systems, paths starting with null byte are invalid
	invalidPath := "/dev/null/invalid/path/test.yaml"
	opts := generator.Options{
		Output: invalidPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), "failed to write YAML to file", "Error should mention file write failure")
	assert.Empty(t, result, "Result should be empty on error")
}

func TestYAMLGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Arrange - Create a generator with a model that can't be marshalled
	// We'll use a function type which YAML cannot marshal
	gen := generator.NewYAMLGenerator[func()]()

	// Create a function value that will cause marshalling to fail
	var model = func() {}

	opts := generator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.Error(t, err, "Generate should fail when marshalling fails")
	assert.Contains(t, err.Error(), "failed to marshal model to YAML", "Error should mention marshalling failure")
	assert.Empty(t, result, "Result should be empty on error")
}

// createTestModel creates a test model with the given version and enabled state.
func createTestModel(version string, enabled bool) TestModel {
	return TestModel{
		Name:    "test-app",
		Version: version,
		Enabled: enabled,
	}
}

// setupTempFile creates a temporary directory and file path for testing.
func setupTempFile(t *testing.T, filename string) (string, string) {
	t.Helper()
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, filename)

	return tempDir, outputPath
}

// createExistingFile creates a file with existing content for testing.
func createExistingFile(t *testing.T, outputPath, content string) {
	t.Helper()

	err := os.WriteFile(outputPath, []byte(content), 0600)
	require.NoError(t, err, "Setup: create existing file")
}

// assertYAMLContent checks that the YAML result contains expected model data.
func assertYAMLContent(t *testing.T, result string, model TestModel) {
	t.Helper()
	assert.Contains(t, result, "Name: test-app", "YAML should contain name")
	assert.Contains(t, result, "Version: "+model.Version, "YAML should contain version")

	if model.Enabled {
		assert.Contains(t, result, "Enabled: true", "YAML should contain enabled")
	} else {
		assert.Contains(t, result, "Enabled: false", "YAML should contain enabled")
	}
}

// verifyFileContent reads and verifies file content matches expected content.
func verifyFileContent(t *testing.T, tempDir, outputPath, expectedContent string) {
	t.Helper()

	fileContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "File should exist")
	assert.Equal(t, expectedContent, string(fileContent), "File content should match expected")
}

// setupExistingFileTest creates a complete test setup for existing file scenarios.
func setupExistingFileTest(
	t *testing.T,
	version string,
	enabled bool,
	force bool,
) (
	*generator.YAMLGenerator[TestModel],
	TestModel,
	string,
	string,
	string,
	generator.Options,
) {
	t.Helper()

	gen := generator.NewYAMLGenerator[TestModel]()
	model := createTestModel(version, enabled)
	tempDir, outputPath := setupTempFile(t, "existing.yaml")
	existingContent := "# existing content"
	createExistingFile(t, outputPath, existingContent)

	opts := generator.Options{
		Output: outputPath,
		Force:  force,
	}

	return gen, model, tempDir, outputPath, existingContent, opts
}
