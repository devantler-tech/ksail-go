package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewRoleCmd creates the gen role command.
func NewRoleCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "role")
}
