package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewNamespaceCmd creates the gen namespace command.
func NewNamespaceCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for namespace
	generator := kubernetes.NewGenerator("namespace")

	// Use the generator to create the command
	return generator.Generate()
}
