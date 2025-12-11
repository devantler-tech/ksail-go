package gen

import (
	"github.com/spf13/cobra"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

// NewConfigMapCmd creates the gen configmap command.
func NewConfigMapCmd(rt *runtime.Runtime) *cobra.Command {
	return createGenCmd(rt, (*kubectl.Client).CreateConfigMapCmd)
}
