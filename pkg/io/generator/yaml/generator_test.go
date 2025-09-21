package yamlgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewYAMLGenerator[map[string]any]()

	createCluster := func(name string) map[string]any {
		return map[string]any{"name": name}
	}

	assertContent := func(t *testing.T, result, _ string) {
		t.Helper()
		snaps.MatchSnapshot(t, result)
	}

	generatortestutils.RunStandardGeneratorTests(
		t,
		gen,
		createCluster,
		"output.yaml",
		assertContent,
	)
}
