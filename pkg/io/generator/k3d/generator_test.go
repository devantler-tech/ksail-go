package k3dgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewK3dGenerator()
	tests := generatortestutils.GetStandardGenerateTestCases("k3d-config.yaml")

	generatortestutils.TestGenerateCommon(
		t,
		tests,
		createTestCluster,
		gen,
		assertK3dYAML,
		"k3d-config.yaml",
	)
}

func TestGenerateExistingFileNoForce(t *testing.T) {
	t.Parallel()

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

func TestGenerateExistingFileWithForce(t *testing.T) {
	t.Parallel()

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

func TestGenerateFileWriteError(t *testing.T) {
	t.Parallel()

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

func TestGenerateMarshalError(t *testing.T) {
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
	spec := v1alpha1.NewClusterSpec()
	spec.Distribution = v1alpha1.DistributionK3d

	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.APIVersion,
			Kind:       v1alpha1.Kind,
		},
		Metadata: v1alpha1.NewClusterMetadata(name),
		Spec:     spec,
	}
}

// testK3dMarshalError runs a test pattern for K3d generator marshal errors.
func testK3dMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha1.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

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
