package eksgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

// createTestCluster creates a test EKS cluster configuration with the given name and region.
func createTestCluster(name, region string) *v1alpha5.ClusterConfig {
	return &v1alpha5.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eksctl.io/v1alpha5",
			Kind:       "ClusterConfig",
		},
		Metadata: &v1alpha5.ClusterMeta{
			Name:   name,
			Region: region,
		},
	}
}

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewEKSGenerator()

	t.Run("basic generation tests", func(t *testing.T) {
		t.Parallel()

		createClusterFunc := func(name string) *v1alpha5.ClusterConfig {
			return createTestCluster(name, "eu-north-1")
		}

		assertContent := func(t *testing.T, result, _ string) {
			t.Helper()
			snaps.MatchSnapshot(t, result)
		}

		generatortestutils.RunStandardGeneratorTests(
			t,
			gen,
			createClusterFunc,
			"eks.yaml",
			assertContent,
		)
	})

	t.Run("error cases", func(t *testing.T) {
		t.Parallel()

		t.Run("missing metadata", func(t *testing.T) {
			t.Parallel()
			testGenerateMissingMetadata(t, gen)
		})
		t.Run("missing cluster name", func(t *testing.T) {
			t.Parallel()
			testGenerateMissingClusterName(t, gen)
		})
		t.Run("missing cluster region", func(t *testing.T) {
			t.Parallel()
			testGenerateMissingClusterRegion(t, gen)
		})
	})
}

func testGenerateMissingMetadata(t *testing.T, gen *generator.EKSGenerator) {
	t.Helper()

	cluster := &v1alpha5.ClusterConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "eksctl.io/v1alpha5",
			Kind:       "ClusterConfig",
		},
	}
	_, err := gen.Generate(cluster, yamlgenerator.Options{})
	require.Error(t, err)
	require.Equal(t, generator.ErrClusterMetadataRequired, err)
}

func testGenerateMissingClusterName(t *testing.T, gen *generator.EKSGenerator) {
	t.Helper()

	cluster := createTestCluster("", "eu-north-1")
	_, err := gen.Generate(cluster, yamlgenerator.Options{})
	require.Error(t, err)
	require.Equal(t, generator.ErrClusterNameRequired, err)
}

func testGenerateMissingClusterRegion(t *testing.T, gen *generator.EKSGenerator) {
	t.Helper()

	cluster := createTestCluster("minimal", "")
	_, err := gen.Generate(cluster, yamlgenerator.Options{})
	require.Error(t, err)
	require.Equal(t, generator.ErrClusterRegionRequired, err)
}
