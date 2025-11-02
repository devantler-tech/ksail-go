package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewClusterRoleBindingCmd creates the gen clusterrolebinding command.
func NewClusterRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for clusterrolebinding
	generator := kubernetes.NewGenerator("clusterrolebinding")

	// Use the generator to create the command
	return generator.Generate()
}
