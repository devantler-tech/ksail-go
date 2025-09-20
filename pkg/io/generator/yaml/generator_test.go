package yamlgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]interface{}]()
	data := map[string]interface{}{"name": "minimal"}
	result, err := gen.Generate(data, generator.Options{})
	require.NoError(t, err)
	require.NotEmpty(t, result)
	snaps.MatchSnapshot(t, result)
}
