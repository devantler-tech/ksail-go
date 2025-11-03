package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewDeploymentCmd creates the gen deployment command.
func NewDeploymentCmd(_ *runtime.Runtime) *cobra.Command {
	return kubectl.MustNewCommand((*kubectl.Client).NewDeploymentCmd)
}
