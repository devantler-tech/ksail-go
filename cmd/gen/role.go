package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewRoleCmd creates the gen role command.
func NewRoleCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewRoleGenerator()

	return generator.Generate()
}
