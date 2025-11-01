package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewTokenCmd creates the gen token command.
func NewTokenCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "token")
}
