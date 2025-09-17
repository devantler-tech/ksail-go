package kustomizationgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	tests := generatortestutils.GetStandardGenerateTestCases("kustomization.yaml")

	generatortestutils.TestGenerateCommon(
		t,
		tests,
		func(name string) *v1alpha1.Cluster {
			cluster := createTestCluster(name)

			return cluster
		},
		generator.NewKustomizationGenerator(createTestCluster("default")),
		func(t *testing.T, result, _ string) {
			t.Helper()
			assertKustomizationYAML(t, result)
		},
		"kustomization.yaml",
	)
}

func TestGenerateExistingFileNoForce(t *testing.T) {
	t.Parallel()

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

func TestGenerateExistingFileWithForce(t *testing.T) {
	t.Parallel()

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

func TestGenerateFileWriteError(t *testing.T) {
	t.Parallel()

	cluster := createTestCluster("error-cluster")
	gen := generator.NewKustomizationGenerator(cluster)

	// Act & Assert
	generatortestutils.TestFileWriteError(
		t,
		gen,
		cluster,
		"kustomization.yaml",
		"write kustomization",
	)
}

func TestGenerateMarshalError(t *testing.T) {
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

	cfg := createTestCluster("new-generator-cluster")

	gen := generator.NewKustomizationGenerator(cfg)

	require.NotNil(t, gen)
	assert.Equal(t, cfg, gen.KSailConfig)
	assert.NotNil(t, gen.Marshaller)
}

// createTestCluster creates a minimal test cluster configuration.
func createTestCluster(name string) *v1alpha1.Cluster {
	return &v1alpha1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.Kind,
			APIVersion: v1alpha1.APIVersion,
		},
		Metadata: metav1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.Spec{
			Distribution: v1alpha1.DistributionKind,
		},
	}
}

// assertKustomizationYAML ensures the generated YAML contains the expected boilerplate.
func assertKustomizationYAML(t *testing.T, result string) {
	t.Helper()
	assert.Contains(
		t,
		result,
		"apiVersion: kustomize.config.k8s.io/v1beta1",
		"YAML should contain API version",
	)
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

	gen := generator.NewKustomizationGenerator(createTestCluster("marshal-error-cluster"))
	gen.Marshaller = generatortestutils.MarshalFailer[*ktypes.Kustomization]{
		Marshaller: nil,
	}
	cluster := createCluster("marshal-error-cluster")

	// Act & Assert
	generatortestutils.TestGeneratorMarshalError[*v1alpha1.Cluster, *ktypes.Kustomization](
		t,
		gen,
		cluster,
		expectedErrorContains,
	)
}
