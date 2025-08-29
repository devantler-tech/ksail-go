package kustomizationgenerator_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

var errBoom = errors.New("boom")

// marshalFailer overrides only Marshal to fail; other methods are satisfied via embedding.
type marshalFailer struct {
	marshaller.Marshaller[*ktypes.Kustomization]
}

func (m marshalFailer) Marshal(_ *ktypes.Kustomization) (string, error) {
	return "", errBoom
}

func TestKustomizationGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("test-cluster")
	gen := generator.NewKustomizationGenerator(cfg)
	opts := generator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKustomizationYAML(t, result)
}

func TestKustomizationGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("file-cluster")
	gen := generator.NewKustomizationGenerator(cfg)
	tempDir := t.TempDir()
	opts := generator.Options{
		Output: tempDir,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKustomizationYAML(t, result)

	// Verify file was written
	expectedPath := filepath.Join(tempDir, "kustomization.yaml")
	assertFileEquals(t, tempDir, expectedPath, result)
}

func TestKustomizationGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("existing-no-force")
	gen := generator.NewKustomizationGenerator(cfg)
	tempDir, outputPath, existingContent := setupExistingFile(t)

	opts := generator.Options{
		Output: tempDir,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKustomizationYAML(t, result)

	// Verify existing file was NOT overwritten
	assertFileEquals(t, tempDir, outputPath, existingContent)
}

func TestKustomizationGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("existing-with-force")
	gen := generator.NewKustomizationGenerator(cfg)
	tempDir, outputPath, existingContent := setupExistingFile(t)

	opts := generator.Options{
		Output: tempDir,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKustomizationYAML(t, result)

	// Verify existing file WAS overwritten
	assertFileEquals(t, tempDir, outputPath, result)
	assert.NotEqual(t, existingContent, result, "Old content should be replaced when Force=true")
}

func TestKustomizationGenerator_Generate_DirectoryCreationError(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("dir-error-cluster")
	gen := generator.NewKustomizationGenerator(cfg)
	// Use an invalid path to force directory creation error
	invalidPath := "/dev/null/invalid"
	opts := generator.Options{
		Output: invalidPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(opts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create kustomization dir")
	assert.Empty(t, result)
}

func TestKustomizationGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("marshal-error-cluster")
	gen := generator.NewKustomizationGenerator(cfg)
	gen.Marshaller = marshalFailer{
		Marshaller: nil,
	}
	opts := generator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(opts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal kustomization")
	assert.Empty(t, result)
}

func TestNewKustomizationGenerator(t *testing.T) {
	t.Parallel()

	// Arrange
	cfg := createTestCluster("new-generator-cluster")

	// Act
	gen := generator.NewKustomizationGenerator(cfg)

	// Assert
	require.NotNil(t, gen)
	assert.Equal(t, cfg, gen.KSailConfig)
	assert.NotNil(t, gen.Marshaller)
}

// createTestCluster creates a minimal test cluster configuration.
func createTestCluster(name string) *v1alpha1.Cluster {
	return v1alpha1.NewCluster(
		v1alpha1.WithMetadataName(name),
		v1alpha1.WithSpecDistribution(v1alpha1.DistributionKind),
		v1alpha1.WithSpecContainerEngine(v1alpha1.ContainerEngineDocker),
	)
}

// assertKustomizationYAML ensures the generated YAML contains the expected boilerplate.
func assertKustomizationYAML(t *testing.T, result string) {
	t.Helper()
	assert.Contains(t, result, "apiVersion: kustomize.config.k8s.io/v1beta1", "YAML should contain API version")
	assert.Contains(t, result, "kind: Kustomization", "YAML should contain kind")
}
// assertFileEquals compares the file content with the expected string.
func assertFileEquals(t *testing.T, dir, path, expected string) {
	t.Helper()

	fileContent, err := ioutils.ReadFileSafe(dir, path)

	require.NoError(t, err, "File should exist")
	assert.Equal(t, expected, string(fileContent))
}

// setupExistingFile creates a temporary directory and an existing kustomization.yaml file
// with default placeholder content, returning the directory, file path, and content string.
func setupExistingFile(t *testing.T) (string, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "kustomization.yaml")
	existingContent := "# existing content"
	err := os.WriteFile(outputPath, []byte(existingContent), 0o600)
	require.NoError(t, err, "Setup: create existing file")

	return tempDir, outputPath, existingContent
}