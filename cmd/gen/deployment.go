package gen

import (
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/devantler-tech/ksail-go/pkg/gen/kubernetes"
	"github.com/spf13/cobra"
)

// NewDeploymentCmd creates the gen deployment command.
func NewDeploymentCmd(_ *runtime.Runtime) *cobra.Command {
	generator := kubernetes.NewDeploymentGenerator()

	return generator.Command()
}
