package kindgenerator_test

import (
	"path/filepath"
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/io"
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
		testGenerateWithoutOutput(t, gen, cluster)
	})
	t.Run("with output file", func(t *testing.T) {
		t.Parallel()
		testGenerateWithOutput(t, gen, cluster)
	})
	t.Run("with output file and force overwrite", func(t *testing.T) {
		t.Parallel()
		testGenerateWithForceOverwrite(t, gen, cluster)
	})
}

func testGenerateWithoutOutput(t *testing.T, gen *generator.KindGenerator,
	cluster *kindv1alpha4.Cluster,
) {
	t.Helper()

	result, err := gen.Generate(cluster, yamlgenerator.Options{})
	require.NoError(t, err)
	require.NotEmpty(t, result)
	snaps.MatchSnapshot(t, result)
}

func testGenerateWithOutput(t *testing.T, gen *generator.KindGenerator,
	cluster *kindv1alpha4.Cluster,
) {
	t.Helper()

	tempDir := t.TempDir()

	outputFile := filepath.Join(tempDir, "kind.yaml")
	result, err := gen.Generate(cluster, yamlgenerator.Options{
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
}

func testGenerateWithForceOverwrite(t *testing.T, gen *generator.KindGenerator,
	cluster *kindv1alpha4.Cluster,
) {
	t.Helper()

	tempDir := t.TempDir()

	outputFile := filepath.Join(tempDir, "kind.yaml")

	// Create file first
	err := io.WriteFileSafe("existing content", tempDir, outputFile, true)
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
	content, err := io.ReadFileSafe(tempDir, outputFile)
	require.NoError(t, err)
	snaps.MatchSnapshot(t, string(content))
}
