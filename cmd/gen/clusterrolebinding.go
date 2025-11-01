package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewClusterRoleBindingCmd creates the gen clusterrolebinding command.
func NewClusterRoleBindingCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "clusterrolebinding")
}
