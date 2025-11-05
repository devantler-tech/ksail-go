package fluxgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen/sync"
)

// SyncOptions defines options for the Flux sync generator.
type SyncOptions struct {
	sync.Options

	Output string // Output file path; if empty, only returns YAML without writing
	Force  bool   // Force overwrite existing files
}

// SyncGenerator generates GitRepository and Kustomization resources for Flux.
type SyncGenerator struct{}

// NewSyncGenerator creates a new SyncGenerator instance.
func NewSyncGenerator() *SyncGenerator {
	return &SyncGenerator{}
}

// Generate creates GitRepository and Kustomization resources and optionally writes to file.
func (g *SyncGenerator) Generate(_ any, opts SyncOptions) (string, error) {
	// Generate the manifest using flux2's sync package
	manifest, err := sync.Generate(opts.Options)
	if err != nil {
		return "", fmt.Errorf("failed to generate Flux sync manifest: %w", err)
	}

	// Write to file if output path is specified
	if opts.Output != "" {
		result, err := io.TryWriteFile(manifest.Content, opts.Output, opts.Force)
		if err != nil {
			return "", fmt.Errorf("failed to write manifest to file: %w", err)
		}

		return result, nil
	}

	return manifest.Content, nil
}
