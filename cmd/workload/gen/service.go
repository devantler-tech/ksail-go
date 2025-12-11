package gen

import (
	"github.com/spf13/cobra"

	"github.com/devantler-tech/ksail-go/pkg/client/kubectl"
	runtime "github.com/devantler-tech/ksail-go/pkg/di"
)

// NewServiceCmd creates the gen service command.
func NewServiceCmd(rt *runtime.Runtime) *cobra.Command {
	return createGenCmd(rt, (*kubectl.Client).CreateServiceCmd)
}
