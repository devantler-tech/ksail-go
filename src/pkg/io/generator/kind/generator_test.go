package kindgenerator_test

import (
	"testing"

	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	kindv1alpha4 "sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewKindGenerator()

	createCluster := func(_ string) *kindv1alpha4.Cluster {
		return &kindv1alpha4.Cluster{
			TypeMeta: kindv1alpha4.TypeMeta{
				APIVersion: "kind.x-k8s.io/v1alpha4",
				Kind:       "Cluster",
			},
		}
	}

	assertContent := func(t *testing.T, result, _ string) {
		t.Helper()
		snaps.MatchSnapshot(t, result)
	}

	generatortestutils.RunStandardGeneratorTests(t, gen, createCluster, "kind.yaml", assertContent)
}
