package gen

import (
	"github.com/spf13/cobra"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

// NewPriorityClassCmd creates the gen priorityclass command.
func NewPriorityClassCmd(rt *runtime.Runtime) *cobra.Command {
	return createGenCmd(rt, (*kubectl.Client).CreatePriorityClassCmd)
}
