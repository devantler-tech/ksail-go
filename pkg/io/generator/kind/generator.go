// Package kindgenerator provides utilities for generating kind cluster configurations.
package kindgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
)

// KindGenerator generates a kind Cluster YAML.
type KindGenerator struct {
	io.FileWriter

	Marshaller marshaller.Marshaller[*v1alpha4.Cluster]
}

// NewKindGenerator creates a new KindGenerator for the given cluster configuration.
func NewKindGenerator() *KindGenerator {
	m := yamlmarshaller.NewMarshaller[*v1alpha4.Cluster]()

	return &KindGenerator{
		FileWriter: io.FileWriter{},
		Marshaller: m,
	}
}

// Generate creates a kind cluster YAML configuration and writes it to the specified output.
func (g *KindGenerator) Generate(cfg *v1alpha4.Cluster, opts yamlgenerator.Options) (string, error) {
	v1alpha4.SetDefaultsCluster(cfg)

	out, err := g.Marshaller.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal kind config: %w", err)
	}

	// write to file if output path is specified
	if opts.Output != "" {
		result, err := g.TryWrite(out, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("write kind config: %w", err)
		}

		return result, nil
	}

	return out, nil
}
