package gen

import (
	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
	"github.com/spf13/cobra"
)

// NewNamespaceCmd creates the gen namespace command.
func NewNamespaceCmd(_ *runtime.Runtime) *cobra.Command {
	return kubectl.MustNewCommand((*kubectl.Client).NewNamespaceCmd)
}
