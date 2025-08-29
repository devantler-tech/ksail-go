package k3dgenerator_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	ioutils "github.com/devantler-tech/ksail-go/pkg/io"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var errBoom = errors.New("boom")

func TestK3dGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("test-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertK3dYAML(t, result, "test-cluster")
}

func TestK3dGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("file-cluster")
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "k3d-config.yaml")
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertK3dYAML(t, result, "file-cluster")

	// Verify file was written
	assertFileEquals(t, tempDir, outputPath, result)
}

func TestK3dGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("existing-cluster")
	tempDir, outputPath, existingContent := setupExistingFile(t)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertK3dYAML(t, result, "existing-cluster")

	// Verify file was NOT overwritten
	assertFileEquals(t, tempDir, outputPath, existingContent)
}

func TestK3dGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("force-cluster")
	tempDir, outputPath, existingContent := setupExistingFile(t)
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertK3dYAML(t, result, "force-cluster")

	// Verify file was overwritten
	assertFileEquals(t, tempDir, outputPath, result)

	// Additional check: ensure old content was replaced
	fileContent, err := ioutils.ReadFileSafe(tempDir, outputPath)
	require.NoError(t, err, "File should exist")
	assert.NotEqual(t, existingContent, string(fileContent), "Old content should be replaced")
}

func TestK3dGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("write-error-cluster")
	// Use a path that will cause a write error (non-existent directory)
	invalidPath := "/non/existent/directory/k3d-config.yaml"
	opts := yamlgenerator.Options{
		Output: invalidPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), "write k3d config", "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}

// marshalFailer overrides only Marshal to fail; other methods are satisfied via embedding.
type marshalFailer struct {
	marshaller.Marshaller[*v1alpha5.SimpleConfig]
}

func (m marshalFailer) Marshal(_ *v1alpha5.SimpleConfig) (string, error) {
	return "", errBoom
}

func TestK3dGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	gen.Marshaller = marshalFailer{
		Marshaller: nil,
	}
	cluster := createTestCluster("marshal-error-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal k3d config")
	assert.Empty(t, result)
}

// createTestCluster creates a minimal test cluster configuration.
func createTestCluster(name string) *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Metadata: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionK3d,
		},
	}
}

// assertK3dYAML ensures the generated YAML contains the expected boilerplate and cluster name.
func assertK3dYAML(t *testing.T, result string, clusterName string) {
	t.Helper()
	assert.Contains(t, result, "apiVersion: k3d.io/v1alpha5", "YAML should contain API version")
	assert.Contains(t, result, "kind: Simple", "YAML should contain kind")
	assert.Contains(t, result, "name: "+clusterName, "YAML should contain cluster name")
}

// assertFileEquals compares the file content with the expected string.
func assertFileEquals(t *testing.T, dir, path, expected string) {
	t.Helper()

	fileContent, err := ioutils.ReadFileSafe(dir, path)

	require.NoError(t, err, "File should exist")
	assert.Equal(t, expected, string(fileContent))
}

// setupExistingFile creates a temporary directory and an existing k3d config file
// with default placeholder content, returning the directory, file path, and content string.
func setupExistingFile(t *testing.T) (string, string, string) {
	t.Helper()

	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "k3d-config.yaml")
	existingContent := "# existing content"
	err := os.WriteFile(outputPath, []byte(existingContent), 0o600)
	require.NoError(t, err, "Setup: create existing file")

	return tempDir, outputPath, existingContent
}