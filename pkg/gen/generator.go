package gen

import (
	"github.com/spf13/cobra"
)

// Generator is the interface for generating Kubernetes resource manifests.
type Generator interface {
	// GenerateCommand creates a cobra command that generates a Kubernetes resource manifest.
	// The resourceType parameter specifies which kubectl create subcommand to wrap (e.g., "namespace", "deployment").
	GenerateCommand(resourceType string) *cobra.Command
}
