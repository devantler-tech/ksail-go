package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewClusterRoleCmd creates the gen clusterrole command.
func NewClusterRoleCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewClusterRoleGenerator()

	return generator.Command()
}
