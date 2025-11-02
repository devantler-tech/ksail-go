package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewClusterRoleBindingCmd creates the gen clusterrolebinding command.
func NewClusterRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewClusterRoleBindingGenerator()

	return generator.Generate()
}
