package generator_test

import (
	"os"
	"path/filepath"
	"testing"

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
	model := TestModel{
		Name:    "test-app",
		Version: "1.0.0",
		Enabled: true,
	}
	opts := generator.Options{
		Output: "",
		Force:  false,
	} // No output file specified

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assert.Contains(t, result, "Name: test-app", "YAML should contain name")
	assert.Contains(t, result, "Version: 1.0.0", "YAML should contain version")
	assert.Contains(t, result, "Enabled: true", "YAML should contain enabled")
}

func TestYAMLGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := TestModel{
		Name:    "test-app",
		Version: "2.0.0",
		Enabled: false,
	}

	// Create temporary file
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "test.yaml")
	opts := generator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assert.Contains(t, result, "Name: test-app", "YAML should contain name")
	assert.Contains(t, result, "Version: 2.0.0", "YAML should contain version")
	assert.Contains(t, result, "Enabled: false", "YAML should contain enabled")

	// Verify file was created
	fileContent, err := os.ReadFile(outputPath) //nolint:gosec // test file path is safe
	require.NoError(t, err, "File should be created")
	assert.Equal(t, result, string(fileContent), "File content should match result")
}

func TestYAMLGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := TestModel{
		Name:    "test-app",
		Version: "3.0.0",
		Enabled: true,
	}

	// Create temporary file with existing content
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.yaml")
	existingContent := "# existing content"
	err := os.WriteFile(outputPath, []byte(existingContent), 0600)
	require.NoError(t, err, "Setup: create existing file")

	opts := generator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assert.Contains(t, result, "Name: test-app", "YAML should be generated")

	// Verify file was not overwritten
	fileContent, err := os.ReadFile(outputPath) //nolint:gosec // test file path is safe
	require.NoError(t, err, "File should exist")
	assert.Equal(t, existingContent, string(fileContent), "File should not be overwritten")
}

func TestYAMLGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := TestModel{
		Name:    "test-app",
		Version: "4.0.0",
		Enabled: false,
	}

	// Create temporary file with existing content
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "existing.yaml")
	existingContent := "# existing content"
	err := os.WriteFile(outputPath, []byte(existingContent), 0600)
	require.NoError(t, err, "Setup: create existing file")

	opts := generator.Options{
		Output: outputPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(model, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assert.Contains(t, result, "Name: test-app", "YAML should be generated")

	// Verify file was overwritten
	fileContent, err := os.ReadFile(outputPath) //nolint:gosec // test file path is safe
	require.NoError(t, err, "File should exist")
	assert.Equal(t, result, string(fileContent), "File should be overwritten with new content")
	assert.NotEqual(t, existingContent, string(fileContent), "Old content should be replaced")
}

func TestYAMLGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewYAMLGenerator[TestModel]()
	model := TestModel{
		Name:    "test-app",
		Version: "5.0.0",
		Enabled: true,
	}

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
