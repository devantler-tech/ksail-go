// Package eksgenerator provides utilities for generating EKS cluster configurations.
package eksgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// EKSGenerator generates an EKS ClusterConfig YAML.
type EKSGenerator struct {
	io.FileWriter

	Marshaller marshaller.Marshaller[*v1alpha5.ClusterConfig]
}

// NewEKSGenerator creates and returns a new EKSGenerator instance.
func NewEKSGenerator() *EKSGenerator {
	m := yamlmarshaller.NewMarshaller[*v1alpha5.ClusterConfig]()

	return &EKSGenerator{
		FileWriter: io.FileWriter{},
		Marshaller: m,
	}
}

// Generate creates an EKS cluster YAML configuration and writes it to the specified output.
func (g *EKSGenerator) Generate(cfg *v1alpha5.ClusterConfig, opts yamlgenerator.Options) (string, error) {
	// Ensure TypeMeta is set before applying defaults
	cfg.TypeMeta = v1alpha5.ClusterConfigTypeMeta()

	out, err := g.Marshaller.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal EKS config: %w", err)
	}

	// write to file if output path is specified
	if opts.Output != "" {
		result, err := g.TryWrite(out, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("write EKS config: %w", err)
		}

		return result, nil
	}

	return out, nil
}

