package configmanager

import (
	"testing"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
)

func TestAddFlagFromFieldUsesOptionalDescription(t *testing.T) {
	t.Parallel()

	selector := AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		v1alpha1.DistributionKind,
	)
	if selector.Description != "" {
		t.Fatalf("expected empty description, got %q", selector.Description)
	}

	withDesc := AddFlagFromField(
		func(c *v1alpha1.Cluster) any { return &c.Spec.Distribution },
		v1alpha1.DistributionKind,
		"Distribution help",
	)

	if withDesc.Description != "Distribution help" {
		t.Fatalf("expected provided description, got %q", withDesc.Description)
	}
}

func TestDefaultClusterFieldSelectorsProvideDefaults(t *testing.T) {
	t.Parallel()

	selectors := DefaultClusterFieldSelectors()
	if len(selectors) != 2 {
		t.Fatalf("expected two selectors, got %d", len(selectors))
	}

	cluster := v1alpha1.NewCluster()
	for _, selector := range selectors {
		field := selector.Selector(cluster)
		if distribution, ok := field.(*v1alpha1.Distribution); ok {
			if selector.DefaultValue != v1alpha1.DistributionKind {
				t.Fatalf("expected distribution default Kind, got %v", selector.DefaultValue)
			}
			*distribution = v1alpha1.DistributionK3d
			if *distribution != v1alpha1.DistributionK3d {
				t.Fatal("selector did not reference distribution field")
			}
			continue
		}

		pathPtr, ok := field.(*string)
		if !ok {
			t.Fatal("selector did not return supported pointer type")
		}
		if selector.DefaultValue != "kind.yaml" {
			t.Fatalf(
				"expected distribution config default 'kind.yaml', got %v",
				selector.DefaultValue,
			)
		}
		*pathPtr = "custom.yaml"
		if *pathPtr != "custom.yaml" {
			t.Fatal("selector did not reference distribution config field")
		}
	}
}

func TestDefaultContextFieldSelector(t *testing.T) {
	t.Parallel()

	selector := DefaultContextFieldSelector()
	cluster := v1alpha1.NewCluster()
	ptr, ok := selector.Selector(cluster).(*string)
	if !ok {
		t.Fatal("expected selector to return *string")
	}

	if selector.DefaultValue != "kind-kind" {
		t.Fatalf("expected default context 'kind-kind', got %v", selector.DefaultValue)
	}

	*ptr = "custom"
	if cluster.Spec.Connection.Context != "custom" {
		t.Fatal("selector did not reference connection context field")
	}

	if selector.Description == "" {
		t.Fatal("expected description for context selector")
	}
}
