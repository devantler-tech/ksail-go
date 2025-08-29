package eksgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestEKSGenerator_Generate_WithoutFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestCluster("test-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "test-cluster")
}

func TestEKSGenerator_Generate_WithFile(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestCluster("file-cluster")
	tempDir := t.TempDir()
	outputPath := filepath.Join(tempDir, "eks-config.yaml")
	opts := yamlgenerator.Options{
		Output: outputPath,
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "file-cluster")
	
	// Verify file was written
	testutils.AssertFileEquals(t, tempDir, outputPath, result)
}

func TestEKSGenerator_Generate_ExistingFile_NoForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestCluster("existing-no-force")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"eks-config.yaml",
		assertEKSYAML,
		"existing-no-force",
		false,
	)
}

func TestEKSGenerator_Generate_ExistingFile_WithForce(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestCluster("existing-with-force")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"eks-config.yaml",
		assertEKSYAML,
		"existing-with-force",
		true,
	)
}

func TestEKSGenerator_Generate_FileWriteError(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestCluster("error-cluster")

	// Use an invalid file path that will cause a write error
	invalidPath := "/dev/null/invalid/path/eks-config.yaml"
	opts := yamlgenerator.Options{
		Output: invalidPath,
		Force:  true,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.Error(t, err, "Generate should fail when file write fails")
	assert.Contains(t, err.Error(), "write EKS config", "Error should mention write failure")
	assert.Empty(t, result, "Result should be empty on error")
}

func TestEKSGenerator_Generate_MarshalError(t *testing.T) {
	t.Parallel()

	// Act & Assert
	testEKSMarshalError(
		t,
		createTestCluster,
		"marshal EKS config",
	)
}

func TestEKSGenerator_Generate_WithCustomOptions(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestClusterWithOptions("custom-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "custom-cluster")
	
	// Verify custom options are applied
	assert.Contains(t, result, "us-east-1", "YAML should contain custom region")
	assert.Contains(t, result, "t3.medium", "YAML should contain custom instance type")
	assert.Contains(t, result, "\"1.25\"", "YAML should contain custom Kubernetes version")
}

func TestEKSGenerator_Generate_DefaultValues(t *testing.T) {
	t.Parallel()

	// Arrange
	gen := generator.NewEKSGenerator()
	cluster := createTestCluster("default-cluster")
	opts := yamlgenerator.Options{
		Output: "",
		Force:  false,
	}

	// Act
	result, err := gen.Generate(cluster, opts)

	// Assert
	require.NoError(t, err, "Generate should succeed")
	assertEKSYAML(t, result, "default-cluster")
	
	// Verify default values are applied
	assert.Contains(t, result, "us-west-2", "YAML should contain default region")
	assert.Contains(t, result, "m5.large", "YAML should contain default instance type")
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
			Distribution: v1alpha1.DistributionEKS,
		},
	}
}

// createTestClusterWithOptions creates a test cluster with custom EKS options.
func createTestClusterWithOptions(name string) *v1alpha1.Cluster {
	cluster := createTestCluster(name)
	cluster.Spec.Options = v1alpha1.Options{
		EKS: v1alpha1.OptionsEKS{
			AWSRegion:         "us-east-1",
			AWSProfile:        "test-profile",
			NodeType:          "t3.medium",
			MinNodes:          2,
			MaxNodes:          5,
			DesiredNodes:      3,
			KubernetesVersion: "1.25",
		},
	}
	return cluster
}

// assertEKSYAML ensures the generated YAML contains the expected boilerplate and cluster name.
func assertEKSYAML(t *testing.T, result string, clusterName string) {
	t.Helper()
	assert.Contains(t, result, "apiVersion: eksctl.io/v1alpha5", "YAML should contain API version")
	assert.Contains(t, result, "kind: ClusterConfig", "YAML should contain kind")
	assert.Contains(t, result, "name: "+clusterName, "YAML should contain cluster name")
	assert.Contains(t, result, "nodeGroups:", "YAML should contain node groups")
	assert.Contains(t, result, clusterName+"-workers", "YAML should contain node group name")
}

// testEKSMarshalError runs a test pattern for EKS generator marshal errors.
func testEKSMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha1.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

	// Arrange
	gen := generator.NewEKSGenerator()
	gen.Marshaller = generatortestutils.MarshalFailer[*v1alpha5.ClusterConfig]{
		Marshaller: nil,
	}
	cluster := createCluster("marshal-error-cluster")

	// Act & Assert
	generatortestutils.TestGeneratorMarshalError[*v1alpha1.Cluster, *v1alpha5.ClusterConfig](
		t,
		gen,
		cluster,
		expectedErrorContains,
	)
}