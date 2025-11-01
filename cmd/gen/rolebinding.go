package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewRoleBindingCmd creates the gen rolebinding command.
func NewRoleBindingCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "rolebinding")
}
