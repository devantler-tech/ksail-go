// Package kustomizationgenerator provides utilities for generating kustomization.yaml files.
package kustomizationgenerator

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/devantler-tech/ksail-go/pkg/apis/cluster/v1alpha1"
	"github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/devantler-tech/ksail-go/pkg/io/marshaller"
	yamlmarshaller "github.com/devantler-tech/ksail-go/pkg/io/marshaller/yaml"
	ktypes "sigs.k8s.io/kustomize/api/types"
)

// dirPermUserRWXGroupRX is the permission for directories created by the generator.
const dirPermUserRWXGroupRX = 0o750

// Options defines options for kustomization generators when emitting files.
type Options struct {
	Output string // Output directory path; kustomization.yaml will be written to this directory
	Force  bool   // Force overwrite existing files
}

// KustomizationGenerator generates a kustomization.yaml.
type KustomizationGenerator struct {
	io.FileWriter

	KSailConfig *v1alpha1.Cluster
	Marshaller  marshaller.Marshaller[*ktypes.Kustomization]
}

// NewKustomizationGenerator creates and returns a new KustomizationGenerator instance.
func NewKustomizationGenerator(cfg *v1alpha1.Cluster) *KustomizationGenerator {
	m := yamlmarshaller.NewMarshaller[*ktypes.Kustomization]()

	return &KustomizationGenerator{
		FileWriter:  io.FileWriter{},
		KSailConfig: cfg,
		Marshaller:  m,
	}
}

// Generate creates a kustomization.yaml file and writes it to the specified output directory.
func (g *KustomizationGenerator) Generate(opts Options) (string, error) {
	//nolint:exhaustruct // Only basic fields needed for minimal kustomization
	kustomization := ktypes.Kustomization{
		TypeMeta: ktypes.TypeMeta{
			APIVersion: "kustomize.config.k8s.io/v1beta1", 
			Kind:       "Kustomization",
		},
		Resources: []string{},
	}
	
	outputFile := filepath.Join(opts.Output, "kustomization.yaml")

	err := os.MkdirAll(filepath.Dir(outputFile), dirPermUserRWXGroupRX)
	if err != nil {
		return "", fmt.Errorf("create kustomization dir: %w", err)
	}
	
	out, err := g.Marshaller.Marshal(&kustomization)
	if err != nil {
		return "", fmt.Errorf("marshal kustomization: %w", err)
	}
	
	result, err := g.TryWrite(out, outputFile, opts.Force)
	if err != nil {
		return "", fmt.Errorf("write kustomization: %w", err)
	}

	return result, nil
}