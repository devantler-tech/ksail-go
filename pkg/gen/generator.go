package gen

import (
	"github.com/spf13/cobra"
)

// Generator is the interface for generating Kubernetes resource manifests.
type Generator interface {
	// Generate creates a cobra command that generates a Kubernetes resource manifest.
	Generate() *cobra.Command
}
