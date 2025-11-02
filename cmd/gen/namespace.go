package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewNamespaceCmd creates the gen namespace command.
func NewNamespaceCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewNamespaceGenerator()

	return generator.Command()
}
