package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewPodDisruptionBudgetCmd creates the gen poddisruptionbudget command.
func NewPodDisruptionBudgetCmd(runtimeContainer *runtime.Runtime) *cobra.Command {
	return createGenCommand(runtimeContainer, "poddisruptionbudget")
}
