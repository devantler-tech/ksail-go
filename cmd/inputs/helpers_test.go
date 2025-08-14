package inputs_test

import (
	"testing"

	"github.com/devantler-tech/ksail-go/cmd/inputs"
	ksailcluster "github.com/devantler-tech/ksail-go/pkg/apis/v1alpha1/cluster"
)

func TestSetInputsOrFallback(t *testing.T) {
	t.Run("overrides only non-zero CLI inputs", func(t *testing.T) {
		// Arrange: create cluster with defaults
		cluster := ksailcluster.NewCluster()
		original := *cluster // copy for later comparison

		// Set some CLI inputs (simulate flags)
		inputs.Name = "custom-name"
		inputs.Distribution = ksailcluster.DistributionK3d
		inputs.ReconciliationTool = ksailcluster.ReconciliationToolFlux
		// Leave SourceDirectory empty so fallback should remain default
		inputs.SourceDirectory = ""
		// ContainerEngine left zero so fallback remains default

		// Act
		inputs.SetInputsOrFallback(cluster)

		// Assert overridden
		if cluster.Metadata.Name != "custom-name" {
			t.Fatalf("expected name overridden to 'custom-name', got %q", cluster.Metadata.Name)
		}

		if cluster.Spec.Distribution != ksailcluster.DistributionK3d {
			t.Fatalf(
				"expected distribution overridden to %q, got %q",
				ksailcluster.DistributionK3d,
				cluster.Spec.Distribution,
			)
		}

		if cluster.Spec.ReconciliationTool != ksailcluster.ReconciliationToolFlux {
			t.Fatalf(
				"expected reconciliation tool overridden to %q, got %q",
				ksailcluster.ReconciliationToolFlux,
				cluster.Spec.ReconciliationTool,
			)
		}
		// Assert fallbacks preserved
		if cluster.Spec.SourceDirectory != original.Spec.SourceDirectory {
			t.Fatalf(
				"expected source directory fallback %q preserved, got %q",
				original.Spec.SourceDirectory,
				cluster.Spec.SourceDirectory,
			)
		}

		if cluster.Spec.ContainerEngine != original.Spec.ContainerEngine {
			t.Fatalf(
				"expected container engine fallback %q preserved, got %q",
				original.Spec.ContainerEngine,
				cluster.Spec.ContainerEngine,
			)
		}

		// Cleanup mutated global inputs for other tests
		inputs.Name = ""
		inputs.Distribution = ""
		inputs.ReconciliationTool = ""
	})
}
