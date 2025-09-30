// Package scaffolderfactory provides factory functions for creating Scaffolder instances
// with either real implementations or stub implementations for testing.
package scaffolderfactory

import (
	"io"

	"github.com/devantler-tech/ksail-go/integration/stubs"
	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	eksgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/eks"
	k3dgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/k3d"
	kindgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kind"
	kustomizationgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/kustomization"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/scaffolder"
	k3dv1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
	eksv1alpha5 "github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

// NewScaffolder creates a new Scaffolder instance with the provided KSail cluster configuration.
// If useStubs is true, it uses stub implementations instead of real generators.
func NewScaffolder(cfg v1alpha1.Cluster, writer io.Writer, useStubs bool) *scaffolder.Scaffolder {
	if useStubs {
		return newScaffolderWithStubs(cfg, writer)
	}
	return newScaffolderWithRealImplementations(cfg, writer)
}

// newScaffolderWithStubs creates a scaffolder using stub implementations.
func newScaffolderWithStubs(cfg v1alpha1.Cluster, writer io.Writer) *scaffolder.Scaffolder {
	ksailGenerator := stubs.NewGeneratorStub[v1alpha1.Cluster, yamlgenerator.Options]().
		WithResult("# KSail config (stub)\napiVersion: ksail.dev/v1alpha1\nkind: Cluster\n")
	kindGenerator := stubs.NewGeneratorStub[*v1alpha4.Cluster, yamlgenerator.Options]().
		WithResult("# Kind config (stub)\nkind: Cluster\napiVersion: kind.x-k8s.io/v1alpha4\n")
	k3dGenerator := stubs.NewGeneratorStub[*k3dv1alpha5.SimpleConfig, yamlgenerator.Options]().
		WithResult("# K3d config (stub)\napiVersion: k3d.io/v1alpha5\nkind: Simple\n")
	eksGenerator := stubs.NewGeneratorStub[*eksv1alpha5.ClusterConfig, yamlgenerator.Options]().
		WithResult("# EKS config (stub)\napiVersion: eksctl.io/v1alpha5\nkind: ClusterConfig\n")
	kustomizationGenerator := stubs.NewGeneratorStub[*ktypes.Kustomization, yamlgenerator.Options]().
		WithResult("# Kustomization config (stub)\napiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\n")

	return &scaffolder.Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGenerator,
		KindGenerator:          kindGenerator,
		K3dGenerator:           k3dGenerator,
		EKSGenerator:           eksGenerator,
		KustomizationGenerator: kustomizationGenerator,
		Writer:                 writer,
	}
}

// newScaffolderWithRealImplementations creates a scaffolder using real implementations.
func newScaffolderWithRealImplementations(
	cfg v1alpha1.Cluster,
	writer io.Writer,
) *scaffolder.Scaffolder {
	ksailGenerator := yamlgenerator.NewYAMLGenerator[v1alpha1.Cluster]()
	kindGenerator := kindgenerator.NewKindGenerator()
	k3dGenerator := k3dgenerator.NewK3dGenerator()
	eksGenerator := eksgenerator.NewEKSGenerator()
	kustomizationGenerator := kustomizationgenerator.NewKustomizationGenerator()

	return &scaffolder.Scaffolder{
		KSailConfig:            cfg,
		KSailYAMLGenerator:     ksailGenerator,
		KindGenerator:          kindGenerator,
		K3dGenerator:           k3dGenerator,
		EKSGenerator:           eksGenerator,
		KustomizationGenerator: kustomizationGenerator,
		Writer:                 writer,
	}
}
