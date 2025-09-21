package k3dgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewK3dGenerator()

	createCluster := func(_ string) *v1alpha5.SimpleConfig {
		return &v1alpha5.SimpleConfig{}
	}

	assertContent := func(t *testing.T, result, _ string) {
		t.Helper()
		snaps.MatchSnapshot(t, result)
	}

	generatortestutils.RunStandardGeneratorTests(t, gen, createCluster, "k3d.yaml", assertContent)
}
