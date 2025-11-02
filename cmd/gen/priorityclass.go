package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewPriorityClassCmd creates the gen priorityclass command.
func NewPriorityClassCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewPriorityClassGenerator()

	return generator.Command()
}
