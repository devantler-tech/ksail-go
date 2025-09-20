package kustomizationgenerator_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/internal/testutils"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	cluster := &v1alpha1.Cluster{}
	gen := generator.NewKustomizationGenerator(cluster)
	result, err := gen.Generate(cluster, yamlgenerator.Options{})
	require.NoError(t, err)
	require.NotEmpty(t, result)
	snaps.MatchSnapshot(t, result)
}
