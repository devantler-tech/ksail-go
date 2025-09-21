package yamlgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/io"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()
	data := map[string]any{"name": "minimal"}

	t.Run("without output file", func(t *testing.T) {
		t.Parallel()

		result, err := gen.Generate(data, generator.Options{})
		require.NoError(t, err)
		require.NotEmpty(t, result)
		snaps.MatchSnapshot(t, result)
	})

	t.Run("with output file", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		outputFile := filepath.Join(tempDir, "output.yaml")
		result, err := gen.Generate(data, generator.Options{
			Output: outputFile,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Verify file was created
		require.FileExists(t, outputFile)

		// Read file content and match snapshot
		content, err := io.ReadFileSafe(tempDir, outputFile)
		require.NoError(t, err)
		snaps.MatchSnapshot(t, string(content))
	})

	t.Run("with output file and force overwrite", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()

		outputFile := filepath.Join(tempDir, "output.yaml")

		// Create file first
		err := io.WriteFileSafe("existing content", tempDir, outputFile, true)
		require.NoError(t, err)

		result, err := gen.Generate(data, generator.Options{
			Output: outputFile,
			Force:  true,
		})
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Verify file was overwritten
		require.FileExists(t, outputFile)

		// Read file content and match snapshot
		content, err := io.ReadFileSafe(tempDir, outputFile)
		require.NoError(t, err)
		snaps.MatchSnapshot(t, string(content))
	})
}
