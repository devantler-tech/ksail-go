// Package k3dgenerator provides utilities for generating k3d cluster configurations.
package k3dgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/k3d-io/k3d/v5/pkg/config/types"
	v1alpha5 "github.com/k3d-io/k3d/v5/pkg/config/v1alpha5"
)

// K3dGenerator generates a k3d SimpleConfig YAML.
type K3dGenerator struct {
	Marshaller marshaller.Marshaller[*v1alpha5.SimpleConfig]
}

// NewK3dGenerator creates and returns a new K3dGenerator instance.
func NewK3dGenerator() *K3dGenerator {
	m := yamlmarshaller.NewMarshaller[*v1alpha5.SimpleConfig]()

	return &K3dGenerator{
		Marshaller: m,
	}
}

// Generate creates a k3d cluster YAML configuration and writes it to the specified output.
func (g *K3dGenerator) Generate(
	cluster *v1alpha1.Cluster,
	opts yamlgenerator.Options,
) (string, error) {
	cfg := g.buildSimpleConfig(cluster)

	out, err := g.Marshaller.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal k3d config: %w", err)
	}

	// write to file if output path is specified
	if opts.Output != "" {
		result, err := io.TryWriteFile(out, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("write k3d config: %w", err)
		}

		return result, nil
	}

	return out, nil
}

func (g *K3dGenerator) buildSimpleConfig(_ *v1alpha1.Cluster) *v1alpha5.SimpleConfig {
	// Create absolutely minimal configuration with explicit TypeMeta
	//nolint:exhaustruct // We only want TypeMeta here
	cfg := &v1alpha5.SimpleConfig{
		TypeMeta: types.TypeMeta{
			APIVersion: "k3d.io/v1alpha3",
			Kind:       "Simple",
		},
	}

	return cfg
}
