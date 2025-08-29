package kindgenerator_test

import (
	"errors"
	"path/filepath"
	"testing"

	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	"github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

var errBoom = errors.New("boom")

func TestKindGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewKindGenerator()
	cluster := createTestCluster("test-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKindYAML(t, result, "test-cluster")
}

func TestKindGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewKindGenerator()
	cluster := createTestCluster("file-cluster")
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "kind-config.yaml")
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKindYAML(t, result, "file-cluster")

	// Verify file was written
	testutils.AssertFileEquals(t, tempDir, outputPath, result)
}

func TestKindGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewKindGenerator()
	cluster := createTestCluster("existing-no-force")

	// Act & Assert
	testutils.TestExistingFileNoForce(
		t,
		gen,
		cluster,
		"kind-config.yaml",
		assertKindYAML,
		"existing-no-force",
	)
}

func TestKindGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewKindGenerator()
	cluster := createTestCluster("existing-with-force")

	// Act & Assert
	testutils.TestExistingFileWithForce(
		t,
		gen,
		cluster,
		"kind-config.yaml",
		assertKindYAML,
		"existing-with-force",
	)
}

func TestKindGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewKindGenerator()
	cluster := createTestCluster("error-cluster")

	// Use an invalid file path that will cause a write error
	invalidPath := "/dev/null/invalid/path/kind-config.yaml"
	opts := yamlgenerator.Options{
		Output: invalidPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), "write kind config", "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}

// marshalFailer overrides only Marshal to fail; other methods are satisfied via embedding.
type marshalFailer struct {
	marshaller.Marshaller[*v1alpha4.Cluster]
}

func (m marshalFailer) Marshal(_ *v1alpha4.Cluster) (string, error) {
	return "", errBoom
}

func TestKindGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewKindGenerator()
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
	assert.Contains(t, err.Error(), "marshal kind config")
	assert.Empty(t, result)
}

// createTestCluster creates a minimal test cluster configuration.
func createTestCluster(name string) *v1alpha4.Cluster {
	cluster := &v1alpha4.Cluster{
		Name: name,
	}

	// Add a minimal control plane node to ensure kind processes the cluster correctly
	var node v1alpha4.Node

	node.Role = v1alpha4.ControlPlaneRole
	cluster.Nodes = append(cluster.Nodes, node)

	return cluster
}

// assertKindYAML ensures the generated YAML contains the expected boilerplate and cluster name.
func assertKindYAML(t *testing.T, result string, clusterName string) {
	t.Helper()
	assert.Contains(t, result, "apiVersion: kind.x-k8s.io/v1alpha4", "YAML should contain API version")
	assert.Contains(t, result, "kind: Cluster", "YAML should contain kind")
	assert.Contains(t, result, "name: "+clusterName, "YAML should contain cluster name")
}
