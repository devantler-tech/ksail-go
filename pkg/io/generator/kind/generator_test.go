package kindgenerator_test

import (
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
	result, err := gen.Generate(cluster, yamlgenerator.Options{})
	require.NoError(t, err)
	require.NotEmpty(t, result)
	snaps.MatchSnapshot(t, result)
}
