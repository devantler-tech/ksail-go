// Package kustomizationgenerator provides utilities for generating kustomization.yaml files.
package kustomizationgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

// KustomizationGenerator generates a kustomization.yaml.
type KustomizationGenerator struct {
	KSailConfig *v1alpha1.Cluster
	Marshaller  marshaller.Marshaller[*ktypes.Kustomization]
}

// NewKustomizationGenerator creates and returns a new KustomizationGenerator instance.
func NewKustomizationGenerator(cfg *v1alpha1.Cluster) *KustomizationGenerator {
	m := yamlmarshaller.NewMarshaller[*ktypes.Kustomization]()

	return &KustomizationGenerator{
		KSailConfig: cfg,
		Marshaller:  m,
	}
}

// Generate creates a kustomization.yaml file and writes it to the specified output file path.
func (g *KustomizationGenerator) Generate(
	_ *v1alpha1.Cluster,
	opts yamlgenerator.Options,
) (string, error) {
	//nolint:exhaustruct // Only basic fields needed for minimal kustomization
	kustomization := ktypes.Kustomization{
		TypeMeta: ktypes.TypeMeta{
			APIVersion: "kustomize.config.k8s.io/v1beta1",
			Kind:       "Kustomization",
		},
		Resources: []string{},
	}

	out, err := g.Marshaller.Marshal(&kustomization)
	if err != nil {
		return "", fmt.Errorf("marshal kustomization: %w", err)
	}

	// If no output file specified, just return the YAML
	if opts.Output == "" {
		return out, nil
	}

	result, err := io.TryWriteFile(out, opts.Output, opts.Force)
	if err != nil {
		return "", fmt.Errorf("write kustomization: %w", err)
	}

	return result, nil
}
