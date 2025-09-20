// Package eksgenerator provides utilities for generating EKS cluster configurations.
package eksgenerator

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/devantler-tech/ksail-go/pkg/io"
	yamlgenerator "github.com/devantler-tech/ksail-go/pkg/io/generator/yaml"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	"github.com/weaveworks/eksctl/pkg/apis/eksctl.io/v1alpha5"
)

// Static errors for validation.
var (
	ErrClusterMetadataRequired = errors.New("cluster metadata is required")
	ErrClusterNameRequired     = errors.New("cluster name is required")
	ErrClusterRegionRequired   = errors.New("cluster region is required")
)

// EKSGenerator generates an EKS ClusterConfig YAML.
type EKSGenerator struct {
	Marshaller marshaller.Marshaller[*v1alpha5.ClusterConfig]
}

// NewEKSGenerator creates and returns a new EKSGenerator instance.
func NewEKSGenerator() *EKSGenerator {
	m := yamlmarshaller.NewMarshaller[*v1alpha5.ClusterConfig]()

	return &EKSGenerator{
		Marshaller: m,
	}
}

// Generate creates an EKS cluster YAML configuration and writes it to the specified output.
func (g *EKSGenerator) Generate(
	cfg *v1alpha5.ClusterConfig,
	opts yamlgenerator.Options,
) (string, error) {
	// Ensure TypeMeta is set before applying defaults
	cfg.TypeMeta = v1alpha5.ClusterConfigTypeMeta()

	// Basic validation - check required fields
	if cfg.Metadata == nil {
		return "", ErrClusterMetadataRequired
	}

	if cfg.Metadata.Name == "" {
		return "", ErrClusterNameRequired
	}

	if cfg.Metadata.Region == "" {
		return "", ErrClusterRegionRequired
	}

	out, err := g.Marshaller.Marshal(cfg)
	if err != nil {
		return "", fmt.Errorf("marshal EKS config: %w", err)
	}

	// Remove empty addonsConfig line if it's just an empty object.
	// It has to be done post-marshalling as the AddonsConfig is not a pointer.
	out = regexp.MustCompile(`(?m)^addonsConfig: \{\}\s*\n`).ReplaceAllString(out, "")

	// write to file if output path is specified
	if opts.Output != "" {
		result, err := io.TryWriteFile(out, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("write EKS config: %w", err)
		}

		return result, nil
	}

	return out, nil
}
