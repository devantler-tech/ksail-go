package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewRoleBindingCmd creates the gen rolebinding command.
func NewRoleBindingCmd(_ *runtime.Runtime) *cobra.Command {
	return kubectl.MustNewCommand((*kubectl.Client).NewRoleBindingCmd)
}
