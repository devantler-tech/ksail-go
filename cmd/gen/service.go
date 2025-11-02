package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewServiceCmd creates the gen service command.
func NewServiceCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewServiceGenerator()

	return generator.Command()
}
