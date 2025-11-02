package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewPriorityClassCmd creates the gen priorityclass command.
func NewPriorityClassCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for priorityclass
	generator := kubernetes.NewGenerator("priorityclass")

	// Use the generator to create the command
	return generator.Generate()
}
