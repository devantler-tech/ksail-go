package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewRoleBindingCmd creates the gen rolebinding command.
func NewRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for rolebinding
	generator := kubernetes.NewGenerator("rolebinding")

	// Use the generator to create the command
	return generator.Generate()
}
