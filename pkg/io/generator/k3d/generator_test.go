package k3dgenerator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewK3dGenerator()
	cluster := &v1alpha5.SimpleConfig{}

	t.Run("without output file", func(t *testing.T) {
		t.Parallel()

		result, err := gen.Generate(cluster, yamlgenerator.Options{})
		require.NoError(t, err)
		require.NotEmpty(t, result)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("with output file", func(t *testing.T) {
		t.Parallel()

		tempDir, err := os.MkdirTemp("", "k3d-generator-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "k3d.yaml")
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

		tempDir, err := os.MkdirTemp("", "k3d-generator-test-*")
		require.NoError(t, err)

		defer os.RemoveAll(tempDir)

		outputFile := filepath.Join(tempDir, "k3d.yaml")

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
