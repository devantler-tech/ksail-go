package eksgenerator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
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

	t.Run("successful generation without output file", func(t *testing.T) {
		t.Parallel()

		cluster := createTestCluster("minimal", "eu-north-1")
		result, err := gen.Generate(cluster, yamlgenerator.Options{})
		require.NoError(t, err)
		require.NotEmpty(t, result)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("successful generation with output file", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "eks-generator-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "eks.yaml")
		cluster := createTestCluster("minimal", "eu-north-1")
		result, err := gen.Generate(cluster, yamlgenerator.Options{
			Output: outputFile,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Verify file was created
		require.FileExists(t, outputFile)

		// Read file content and match snapshot
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		snaps.MatchSnapshot(t, string(content))
	})

	t.Run("successful generation with output file and force overwrite", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "eks-generator-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "eks.yaml")

		// Create file first
		err = os.WriteFile(outputFile, []byte("existing content"), 0o644)
		require.NoError(t, err)

		cluster := createTestCluster("minimal", "eu-north-1")
		result, err := gen.Generate(cluster, yamlgenerator.Options{
			Output: outputFile,
			Force:  true,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Verify file was overwritten
		require.FileExists(t, outputFile)

		// Read file content and match snapshot
		content, err := os.ReadFile(outputFile)
		require.NoError(t, err)
		snaps.MatchSnapshot(t, string(content))
	})

	t.Run("missing metadata", func(t *testing.T) {
		t.Parallel()

		cluster := &v1alpha5.ClusterConfig{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "eksctl.io/v1alpha5",
				Kind:       "ClusterConfig",
			},
		}
		_, err := gen.Generate(cluster, yamlgenerator.Options{})
		require.Error(t, err)
		require.Equal(t, generator.ErrClusterMetadataRequired, err)
	})

	t.Run("missing cluster name", func(t *testing.T) {
		t.Parallel()

		cluster := createTestCluster("", "eu-north-1")
		_, err := gen.Generate(cluster, yamlgenerator.Options{})
		require.Error(t, err)
		require.Equal(t, generator.ErrClusterNameRequired, err)
	})

	t.Run("missing cluster region", func(t *testing.T) {
		t.Parallel()

		cluster := createTestCluster("minimal", "")
		_, err := gen.Generate(cluster, yamlgenerator.Options{})
		require.Error(t, err)
		require.Equal(t, generator.ErrClusterRegionRequired, err)
	})
}
