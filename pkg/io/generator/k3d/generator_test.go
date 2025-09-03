package k3dgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	clustertestutils "github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	k3dtestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d/testutils"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
	testutils.AssertFileEquals(t, tempDir, outputPath, result)
}

func TestK3dGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("existing-cluster")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"k3d-config.yaml",
		assertK3dYAML,
		"existing-cluster",
		false,
	)
}

func TestK3dGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("force-cluster")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"k3d-config.yaml",
		assertK3dYAML,
		"force-cluster",
		true,
	)
}

func TestK3dGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewK3dGenerator()
	cluster := createTestCluster("write-error-cluster")

	// Act & Assert
	generatortestutils.TestFileWriteError(
		t,
		gen,
		cluster,
		"k3d-config.yaml",
		"write k3d config",
	)
}

func TestK3dGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Act & Assert
	testK3dMarshalError(
		t,
		createTestCluster,
		"marshal k3d config",
	)
}

// createTestCluster creates a minimal test cluster configuration.
func createTestCluster(name string) *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Metadata: clustertestutils.CreateDefaultObjectMeta(name),
		Spec:     k3dtestutils.CreateDefaultK3dSpec(),
	}
}

// testK3dMarshalError runs a test pattern for K3d generator marshal errors.
func testK3dMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha1.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange
	gen := generator.NewK3dGenerator()
	gen.Marshaller = generatortestutils.MarshalFailer[*v1alpha5.SimpleConfig]{
		Marshaller: nil,
	}
	cluster := createCluster("marshal-error-cluster")

	// Act & Assert
	generatortestutils.TestGeneratorMarshalError[*v1alpha1.Cluster, *v1alpha5.SimpleConfig](
		t,
		gen,
		cluster,
		expectedErrorContains,
	)
}

// assertK3dYAML ensures the generated YAML contains the expected boilerplate and cluster name.
func assertK3dYAML(t *testing.T, result string, clusterName string) {
	t.Helper()
	assert.Contains(t, result, "apiVersion: k3d.io/v1alpha5", "YAML should contain API version")
	assert.Contains(t, result, "kind: Simple", "YAML should contain kind")
	assert.Contains(t, result, "name: "+clusterName, "YAML should contain cluster name")
}
