package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewPodDisruptionBudgetCmd creates the gen poddisruptionbudget command.
func NewPodDisruptionBudgetCmd(_ *runtime.Runtime) *cobra.Command {
	// Create a generator for poddisruptionbudget
	generator := kubernetes.NewGenerator("poddisruptionbudget")

	// Use the generator to create the command
	return generator.Generate()
}
