package gen

import (
	"github.com/spf13/cobra"
)

// Generator is the interface for generating Kubernetes resource manifests.
type Generator interface {
	// Command returns the cobra command for generating the Kubernetes resource manifest.
	Command() *cobra.Command
}
