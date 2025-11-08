package kustomizationgenerator_test

import (
	"testing"

	generator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	generatortestutils "github.com/devantler-tech/ksail-go/pkg/io/generator/testutils"
	"github.com/devantler-tech/ksail-go/pkg/testutils"
	"github.com/gkampitakis/go-snaps/snaps"
	"sigs.k8s.io/kustomize/api/types"
)

func TestMain(m *testing.M) { testutils.RunTestMainWithSnapshotCleanup(m) }

func TestGenerate(t *testing.T) {
	t.Parallel()

	gen := generator.NewKustomizationGenerator()

	createCluster := func(_ string) *types.Kustomization {
		return &types.Kustomization{}
	}

	assertContent := func(t *testing.T, result, _ string) {
		t.Helper()
		snaps.MatchSnapshot(t, result)
	}

	generatortestutils.RunStandardGeneratorTests(
		t,
		gen,
		createCluster,
		"kustomization.yaml",
		assertContent,
	)
}
