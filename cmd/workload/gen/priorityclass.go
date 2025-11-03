package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewPriorityClassCmd creates the gen priorityclass command.
func NewPriorityClassCmd(_ *runtime.Runtime) *cobra.Command {
	client := kubectl.DefaultClient()
	cmd, err := client.CreatePriorityClassCmd()
	cobra.CheckErr(err)

	return cmd
}
