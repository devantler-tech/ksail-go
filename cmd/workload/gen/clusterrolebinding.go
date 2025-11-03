package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewClusterRoleBindingCmd creates the gen clusterrolebinding command.
func NewClusterRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	return kubectl.MustNewCommand((*kubectl.Client).NewClusterRoleBindingCmd)
}
