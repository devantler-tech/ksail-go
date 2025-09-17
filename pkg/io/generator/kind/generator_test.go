package kindgenerator_test

import (
	"testing"

	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewKindGenerator()
	tests := generatortestutils.GetStandardGenerateTestCases("kind-config.yaml")

	generatortestutils.TestGenerateCommon(
		t,
		tests,
		createTestCluster,
		gen,
		assertKindYAML,
		"kind-config.yaml",
	)
}

func TestGenerateExistingFileNoForce(t *testing.T) {
	t.Parallel()

	gen := generator.NewKindGenerator()
	cluster := createTestCluster("existing-no-force")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"kind-config.yaml",
		assertKindYAML,
		"existing-no-force",
		false,
	)
}

func TestGenerateExistingFileWithForce(t *testing.T) {
	t.Parallel()

	gen := generator.NewKindGenerator()
	cluster := createTestCluster("existing-with-force")

	// Act & Assert
	generatortestutils.TestExistingFile(
		t,
		gen,
		cluster,
		"kind-config.yaml",
		assertKindYAML,
		"existing-with-force",
		true,
	)
}

func TestGenerateFileWriteError(t *testing.T) {
	t.Parallel()

	gen := generator.NewKindGenerator()
	cluster := createTestCluster("error-cluster")

	// Act & Assert
	generatortestutils.TestFileWriteError(
		t,
		gen,
		cluster,
		"kind-config.yaml",
		"write kind config",
	)
}

func TestGenerateMarshalError(t *testing.T) {
	t.Parallel()

	// Act & Assert
	testKindMarshalError(
		t,
		createTestCluster,
		"marshal kind config",
	)
}

// createTestCluster creates a minimal test cluster configuration.
func createTestCluster(name string) *v1alpha4.Cluster {
	cluster := &v1alpha4.Cluster{
		TypeMeta: v1alpha4.TypeMeta{
			APIVersion: "",
			Kind:       "",
		},
		Name:  name,
		Nodes: nil,
		Networking: v1alpha4.Networking{
			IPFamily:          "",
			APIServerPort:     0,
			APIServerAddress:  "",
			PodSubnet:         "",
			ServiceSubnet:     "",
			DisableDefaultCNI: false,
			KubeProxyMode:     "",
			DNSSearch:         nil,
		},
		FeatureGates:                    nil,
		RuntimeConfig:                   nil,
		KubeadmConfigPatches:            nil,
		KubeadmConfigPatchesJSON6902:    nil,
		ContainerdConfigPatches:         nil,
		ContainerdConfigPatchesJSON6902: nil,
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
	assert.Contains(
		t,
		result,
		"apiVersion: kind.x-k8s.io/v1alpha4",
		"YAML should contain API version",
	)
	assert.Contains(t, result, "kind: Cluster", "YAML should contain kind")
	assert.Contains(t, result, "name: "+clusterName, "YAML should contain cluster name")
}

// testKindMarshalError runs a test pattern for Kind generator marshal errors.
func testKindMarshalError(
	t *testing.T,
	createCluster func(string) *v1alpha4.Cluster,
	expectedErrorContains string,
) {
	t.Helper()

	gen := generator.NewKindGenerator()
	gen.Marshaller = generatortestutils.MarshalFailer[*v1alpha4.Cluster]{
		Marshaller: nil,
	}
	cluster := createCluster("marshal-error-cluster")

	// Act & Assert
	generatortestutils.TestGeneratorMarshalError[*v1alpha4.Cluster, *v1alpha4.Cluster](
		t,
		gen,
		cluster,
		expectedErrorContains,
	)
}
