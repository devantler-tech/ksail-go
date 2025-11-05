package fluxgenerator

import (
	"fmt"

	"github.com/devantler-tech/ksail-go/pkg/io"
	"github.com/fluxcd/flux2/v2/pkg/manifestgen/install"
)

// InstallOptions defines options for the Flux install generator.
type InstallOptions struct {
	install.Options

	Output string // Output file path; if empty, only returns YAML without writing
	Force  bool   // Force overwrite existing files
}

// InstallGenerator generates Flux installation manifests.
type InstallGenerator struct{}

// NewInstallGenerator creates a new InstallGenerator instance.
func NewInstallGenerator() *InstallGenerator {
	return &InstallGenerator{}
}

// Generate creates Flux installation manifests and optionally writes to file.
func (g *InstallGenerator) Generate(_ any, opts InstallOptions) (string, error) {
	// Generate the manifest using flux2's install package
	manifest, err := install.Generate(opts.Options, "")
	if err != nil {
		return "", fmt.Errorf("failed to generate Flux install manifest: %w", err)
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
