package kustomizationgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

func TestKustomizationGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	cluster := createTestCluster("test-cluster")
	gen := generator.NewKustomizationGenerator(cluster)
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKustomizationYAML(t, result)
}

func TestKustomizationGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	cluster := createTestCluster("file-cluster")
	gen := generator.NewKustomizationGenerator(cluster)
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "kustomization.yaml")
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertKustomizationYAML(t, result)

	// Verify file was written
	testutils.AssertFileEquals(t, tempDir, outputPath, result)
}

func TestKustomizationGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	cluster := createTestCluster("existing-no-force")
	gen := generator.NewKustomizationGenerator(cluster)

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"kustomization.yaml",
		assertKustomizationYAMLWithName,
		"existing-no-force",
		false,
	)
}

func TestKustomizationGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	cluster := createTestCluster("existing-with-force")
	gen := generator.NewKustomizationGenerator(cluster)

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"kustomization.yaml",
		assertKustomizationYAMLWithName,
		"existing-with-force",
		true,
	)
}

func TestKustomizationGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	cluster := createTestCluster("error-cluster")
	gen := generator.NewKustomizationGenerator(cluster)

	// Use an invalid file path that will cause a write error
	invalidPath := "/dev/null/invalid/path/kustomization.yaml"
	opts := yamlgenerator.Options{
		Output: invalidPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), "write kustomization", "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}

func TestKustomizationGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Act & Assert
	testKustomizationMarshalError(
		t,
		createTestCluster,
		"marshal kustomization",
	)
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

// assertKustomizationYAMLWithName wraps assertKustomizationYAML for use with testutils.
func assertKustomizationYAMLWithName(t *testing.T, result string, _ string) {
	t.Helper()
	assertKustomizationYAML(t, result)
}

// testKustomizationMarshalError runs a test pattern for Kustomization generator marshal errors.
func testKustomizationMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha1.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange
	cluster := createCluster("marshal-error-cluster")
	gen := generator.NewKustomizationGenerator(cluster)
	gen.Marshaller = generatortestutils.MarshalFailer[*ktypes.Kustomization]{
		Marshaller: nil,
	}

	// Act & Assert
	generatortestutils.TestGeneratorMarshalError[*v1alpha1.Cluster, *ktypes.Kustomization](
		t,
		gen,
		cluster,
		expectedErrorContains,
	)
}