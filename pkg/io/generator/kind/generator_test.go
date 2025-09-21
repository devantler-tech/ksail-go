package kindgenerator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewKindGenerator()
	cluster := &kindv1alpha4.Cluster{
		TypeMeta: kindv1alpha4.TypeMeta{
			APIVersion: "kind.x-k8s.io/v1alpha4",
			Kind:       "Cluster",
		},
	}

	t.Run("without output file", func(t *testing.T) {
		t.Parallel()

		result, err := gen.Generate(cluster, yamlgenerator.Options{})
		require.NoError(t, err)
		require.NotEmpty(t, result)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("with output file", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "kind-generator-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "kind.yaml")
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

	t.Run("with output file and force overwrite", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "kind-generator-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "kind.yaml")

		// Create file first
		err = os.WriteFile(outputFile, []byte("existing content"), 0o644)
		require.NoError(t, err)

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
}
